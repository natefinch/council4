package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerContent lowers one ability body and, when the body carries a reflexive
// "When you do," gate (CR 603.11), repackages the lowered content so the gated
// consequence becomes a reflexive triggered ability whose targets are chosen
// after the enabling action, rather than resolving inline with up-front targets.
// The dispatch itself is text-blind and gate-shape agnostic; the reflexive marker
// travels on the compiled condition, so every lowering path that produces the
// "you may X. When you do, Y" shape is repackaged uniformly here.
func lowerContent(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	content, diagnostic := lowerContentDispatch(cardName, ctx, syntax)
	if diagnostic != nil {
		return content, diagnostic
	}
	if !contentHasReflexiveGate(ctx.content) {
		return content, nil
	}
	if abilityContentHasReflexiveTrigger(content) {
		// An inner lowering frame already repackaged this reflexive gate (the
		// gate condition survived into a recursive call); wrapping it again would
		// double-nest the reflexive trigger, so leave the repackaged content as-is.
		return content, nil
	}
	repackaged, category, ok := repackageReflexiveTrigger(content)
	if !ok {
		return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, category)
	}
	return repackaged, nil
}

// contentHasReflexiveGate reports whether any compiled condition on this content
// is a reflexive "When you do," gate (CR 603.11). The parser marks such a gate
// (ConditionClause.Reflexive) so the compiler and lowering distinguish it from an
// immediate "If you do," rider without inspecting oracle text.
func contentHasReflexiveGate(content compiler.AbilityContent) bool {
	for i := range content.Conditions {
		condition := content.Conditions[i]
		if condition.Reflexive && isResolvingSuccessGate(condition.Predicate) {
			return true
		}
	}
	return false
}

// abilityContentHasReflexiveTrigger reports whether any top-level mode sequence
// already contains a reflexive-trigger instruction. It guards the reflexive
// repackage against double application when a nested lowering frame repackaged
// the same gate.
func abilityContentHasReflexiveTrigger(content game.AbilityContent) bool {
	for i := range content.Modes {
		for j := range content.Modes[i].Sequence {
			if _, ok := content.Modes[i].Sequence[j].Primitive.(game.CreateReflexiveTrigger); ok {
				return true
			}
		}
	}
	return false
}

// repackageReflexiveTrigger rewrites lowered "you may X. When you do, Y" content
// into the reflexive-trigger model. The lowered content resolves the consequence
// Y inline, gated on the enabling action X's published success, with Y's targets
// chosen up front. The reflexive model instead leaves the enabling action inline
// and, when it succeeds, puts a reflexive triggered ability carrying Y on the
// stack (CreateReflexiveTrigger); Y's targets are chosen when the trigger is put
// on the stack, after the enabling action has resolved (CR 603.11).
//
// It repackages only the clean, verified shape: single non-modal content whose
// trailing instructions form a contiguous run gated on one enabling action's
// success, with every target belonging to that gated consequence. Any other shape
// (a targeted enabling action, an "if you don't" else branch, a non-contiguous
// gate, an unvalidatable rewrite) fails closed with a category so the card is
// reported rather than silently mis-modelled.
func repackageReflexiveTrigger(content game.AbilityContent) (game.AbilityContent, string, bool) {
	if content.IsModal() || len(content.Modes) != 1 || len(content.SharedTargets) != 0 {
		return game.AbilityContent{}, "structural — reflexive gate not on single-mode content", false
	}
	mode := content.Modes[0]
	seq := mode.Sequence
	if len(seq) < 2 {
		return game.AbilityContent{}, "structural — reflexive body too short to repackage", false
	}
	// The reflexive consequence is the trailing run of instructions gated on the
	// enabling action's published success. Its key comes from the final
	// instruction's gate; walk back over the contiguous {key, TriTrue} suffix.
	lastGate := seq[len(seq)-1].ResultGate
	if !reflexiveSuccessGate(lastGate) {
		return game.AbilityContent{}, "structural — reflexive consequence is not a trailing success-gated run", false
	}
	key := lastGate.Val.Key
	tailStart := len(seq)
	for i := len(seq) - 1; i >= 0; i-- {
		if reflexiveTailGated(seq[i], key) {
			tailStart = i
			continue
		}
		break
	}
	if tailStart == 0 {
		return game.AbilityContent{}, "structural — reflexive consequence has no enabling action", false
	}
	// The enabling action is the sole publisher of the gate key, sits before the
	// gated tail, and no earlier instruction consumes the key.
	enablingIdx := -1
	for i := 0; i < tailStart; i++ {
		if seq[i].PublishResult == key {
			if enablingIdx != -1 {
				return game.AbilityContent{}, "structural — reflexive gate key published more than once", false
			}
			enablingIdx = i
		}
		if seq[i].ResultGate.Exists && seq[i].ResultGate.Val.Key == key {
			return game.AbilityContent{}, "structural — reflexive gate consumed before its enabling action", false
		}
	}
	if enablingIdx == -1 {
		return game.AbilityContent{}, "structural — reflexive gate has no enabling action", false
	}
	// No consequence instruction may republish the gate key, and each must be
	// gated exactly on {key, TriTrue} (the contiguous-suffix walk guarantees the
	// latter, re-checked so a mixed gate fails closed rather than silently drops).
	for i := tailStart; i < len(seq); i++ {
		if seq[i].PublishResult == key {
			return game.AbilityContent{}, "structural — reflexive consequence republishes the gate key", false
		}
		if !reflexiveTailGated(seq[i], key) {
			return game.AbilityContent{}, "structural — reflexive consequence gate is not uniform", false
		}
	}
	// Build the reflexive ability body: the gated consequence with its result
	// gates cleared (the reflexive trigger fires only when queued, so its body is
	// unconditional), owning all of the content's targets. The targets keep their
	// clause-local indices because the whole target list moves intact, so object-
	// and card-domain numbering is unchanged and no rebasing is needed.
	innerSeq := make([]game.Instruction, 0, len(seq)-tailStart)
	for i := tailStart; i < len(seq); i++ {
		instr := seq[i]
		instr.ResultGate = opt.V[game.InstructionResultGate]{}
		innerSeq = append(innerSeq, instr)
	}
	innerMode := game.Mode{
		Targets:  append([]game.TargetSpec(nil), mode.Targets...),
		Sequence: innerSeq,
	}
	// Reassemble the outer sequence: every instruction up to the consequence,
	// then a single reflexive-trigger instruction gated on the enabling action's
	// success. The outer content no longer targets anything up front.
	outerSeq := make([]game.Instruction, 0, tailStart+1)
	outerSeq = append(outerSeq, seq[:tailStart]...)
	outerSeq = append(outerSeq, game.Instruction{
		Primitive: game.CreateReflexiveTrigger{
			Trigger: game.ReflexiveTriggerDef{Content: innerMode.Ability()},
		},
		ResultGate: opt.Val(game.InstructionResultGate{Key: key, Succeeded: game.TriTrue}),
	})
	outerMode := mode
	outerMode.Targets = nil
	outerMode.Sequence = outerSeq
	repackaged := outerMode.Ability()
	// Validate the rewrite: the outer sequence must reference no targets (all
	// targets moved to the reflexive body), and the nested reflexive body must be
	// self-consistent. A dangling reference (a targeted enabling action) fails
	// closed here rather than emitting silently wrong content.
	if err := game.ValidateInstructionSequence(repackaged.Modes[0].Sequence, repackaged.Modes[0].Targets); err != nil {
		return game.AbilityContent{}, "structural — reflexive repackage did not validate: " + err.Error(), false
	}
	return repackaged, "", true
}

// reflexiveSuccessGate reports whether a result gate is an unqualified
// success gate ({key, TriTrue} with no amount-range qualifier).
func reflexiveSuccessGate(gate opt.V[game.InstructionResultGate]) bool {
	return gate.Exists &&
		gate.Val.Key != "" &&
		gate.Val.Succeeded == game.TriTrue &&
		!gate.Val.AmountRange.Exists
}

// reflexiveTailGated reports whether an instruction belongs to the reflexive
// consequence: gated exactly on the enabling action's success key.
func reflexiveTailGated(instr game.Instruction, key game.ResultKey) bool {
	gate := instr.ResultGate
	return reflexiveSuccessGate(gate) && gate.Val.Key == key
}
