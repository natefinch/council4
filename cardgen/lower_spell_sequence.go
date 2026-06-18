package cardgen

import (
	"fmt"
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func legacyOrderedEffectSequenceExact(effects []compiler.CompiledEffect) bool {
	if len(effects) != 2 {
		return true
	}
	first, second := effects[0], effects[1]
	if first.Kind == compiler.EffectPut && second.Kind == compiler.EffectProliferate {
		return false
	}
	if first.Kind == compiler.EffectModifyPT && second.Kind == compiler.EffectModifyPT &&
		second.Connection == parser.EffectConnectionAnd {
		return false
	}
	if first.Kind == compiler.EffectExile &&
		second.Kind == compiler.EffectReturn &&
		second.DelayedTiming != 0 {
		return referencesBindTo(second.References, compiler.ReferenceBindingPriorInstructionResult, 0)
	}
	return true
}

func lowerOrderedEffectSequence(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	if len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — sequence carries modal options")
	}
	for _, target := range ctx.content.Targets {
		if _, ok := counterAbilityTargetSpec(target); ok {
			return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — counter-spell target")
		}
	}
	// The combined-shape lowerers do not model per-effect conditions; only
	// attempt them when the sequence carries none, so a condition can never be
	// silently dropped.
	if content, ok := lowerCombinedSequenceShapes(cardName, ctx); ok {
		return content, nil
	}
	if !legacyOrderedEffectSequenceExact(ctx.content.Effects) {
		return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — non-exact legacy effect pair")
	}
	// Resolving optionality ("you may X. If you do, Y") is realized by marking
	// the optional effect's instruction Optional + PublishResult and gating the
	// "if you do" effect on that result. planOptionalFlow fails closed unless the
	// optionality forms exactly one supported pair.
	optionalFlow, ok := planOptionalFlow(ctx.content)
	if !ok {
		return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — unsupported resolving optionality")
	}
	// The affirmative "if you do" clause is consumed as the optional-flow gate,
	// not as an ordinary effect-gate condition, so exclude it from the per-effect
	// condition matching (its predicate is not a supported effect-gate predicate).
	gateConditions := optionalFlowGateConditions(ctx.content.Conditions, optionalFlow)
	// Match each condition to the single effect whose clause span contains it and
	// lower it as an effect gate. Fails closed if any condition is not contained
	// in exactly one effect or is not a supported effect-gate condition.
	effectConditions, ok := matchSequenceEffectConditions(ctx.content.Effects, gateConditions)
	if !ok {
		return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — per-effect condition not matched to one clause")
	}
	consumedConditions := 0
	if optionalFlow.enabled {
		consumedConditions++
	}
	var targets []game.TargetSpec
	var sequence []game.Instruction
	consumedTargets := 0
	consumedKeywords := 0
	consumedReferences := 0
	// oracleSpanToGameIdx maps each oracle target's Span to its first index in
	// the accumulated targets slice, recorded when the target is owned (i.e.
	// added as a new game.TargetSpec by a non-shared clause). This index is
	// looked up when an inherited shared-target clause needs to rebase its
	// sequence: the rebase offset equals the start index of the inherited
	// target rather than always 0, which is wrong when earlier effects already
	// contributed target specs before the then-joined group.
	oracleSpanToGameIdx := make(map[shared.Span]int)
	clauseSyntaxes := splitEffectSyntaxes(syntax, ctx.content.Effects)
	for i := range ctx.content.Effects {
		effect := &ctx.content.Effects[i]
		resolvedEffect, clauseAbility := prepareSequenceClause(ctx, optionalFlow, clauseSyntaxes, i)
		effectAbility := contextForEffect(ctx, &resolvedEffect)
		// Per-effect conditions are handled by the sequence gate (effectConditions),
		// not by the individual effect lowerers, so clear the content-level
		// conditions inherited from the parent context before per-effect lowering.
		effectAbility.content.Conditions = nil
		clauseTargets := effect.Targets
		clauseRefs := effect.References
		ownedReferenceCount := len(clauseRefs)
		var inheritedTargets []compiler.CompiledTarget
		if effect.Context == parser.EffectContextPriorSubject {
			inheritedTargets = priorSubjectTargets(ctx.content.Effects, i)
			clauseRefs = append(clauseRefs, priorSubjectReferences(ctx.content.Effects, i)...)
		}
		inheritedTargets = appendReferenceAntecedentTargets(
			inheritedTargets,
			clauseRefs,
			ctx.content.Targets,
			clauseTargets,
		)
		// Three target-handling modes:
		//   allSharedTargets: only inherited, no own — compound-mill "then draws".
		//   mixedTargets:     inherited + own — "then fights target creature" where
		//                     the inherited subject and a new object both appear.
		//   otherwise:        only own (or none) — normal independent effects.
		allSharedTargets := len(inheritedTargets) > 0 && len(clauseTargets) == 0
		mixedTargets := len(inheritedTargets) > 0 && len(clauseTargets) > 0
		switch {
		case allSharedTargets:
			effectAbility.content.Targets = inheritedTargets
		case mixedTargets:
			combined := make([]compiler.CompiledTarget, 0, len(inheritedTargets)+len(clauseTargets))
			combined = append(combined, inheritedTargets...)
			combined = append(combined, clauseTargets...)
			effectAbility.content.Targets = combined
		default:
			effectAbility.content.Targets = clauseTargets
		}
		effectAbility.content.References = clauseRefs
		localReferences, ok := localizeTargetReferences(
			effectAbility.content.References,
			ctx.content.Targets,
			effectAbility.content.Targets,
		)
		if !ok {
			return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — clause reference not localizable")
		}
		effectAbility.content.References = localReferences
		effectAbility.content.Keywords = keywordsWithinSpan(ctx.content.Keywords, effect.ClauseSpan)
		consumedTargets += len(clauseTargets)
		consumedKeywords += len(effectAbility.content.Keywords)
		consumedReferences += ownedReferenceCount
		// Lower the effect through the shared lowerAbilityContent entry point.
		// allSharedTargets: try with inherited targets; if that fails, retry
		//   with targets cleared (e.g. "then proliferate" rejects any target).
		// mixedTargets: inherited+own combined — no fallback (fail-closed).
		// default: straightforward lowering with own targets only.
		var content game.AbilityContent
		var diagnostic *shared.Diagnostic
		if delayedContent, handled, failed := lowerDelayedSequenceClause(ctx.content.Effects, i, effectAbility, sequence); handled {
			if failed {
				return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — delayed-target sacrifice not linkable")
			}
			content = delayedContent
		} else if allSharedTargets {
			content, diagnostic = lowerSequenceClauseContent(cardName, effectAbility.content, effectAbility.optional, &clauseAbility)
			if diagnostic != nil {
				effectAbilityNoTarget := effectAbility
				effectAbilityNoTarget.content.Targets = nil
				content, diagnostic = lowerSequenceClauseContent(cardName, effectAbilityNoTarget.content, effectAbilityNoTarget.optional, &clauseAbility)
			}
		} else {
			content, diagnostic = lowerSequenceClauseContent(cardName, effectAbility.content, effectAbility.optional, &clauseAbility)
		}
		if diagnostic != nil ||
			len(content.SharedTargets) != 0 ||
			content.IsModal() ||
			len(content.Modes) != 1 {
			return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, sequenceClauseCategory(diagnostic))
		}
		mode := content.Modes[0]
		newTargets, ok := applyTargetRemapping(
			mode, allSharedTargets, mixedTargets,
			inheritedTargets, clauseTargets,
			targets, oracleSpanToGameIdx,
		)
		if !ok {
			return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — inherited target not remappable")
		}
		targets = newTargets
		if effectCondition, gated := effectConditions[i]; gated {
			if !applyEffectConditionGate(mode.Sequence, &effectCondition) {
				return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — per-effect condition gate not applicable")
			}
			consumedConditions++
		}
		if optionalFlow.enabled || optionalFlow.bareIndex >= 0 {
			if category, ok := applyOptionalFlowEnvelope(optionalFlow, i, mode.Sequence); !ok {
				return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, category)
			}
		}
		sequence = append(sequence, mode.Sequence...)
	}
	if consumedTargets != len(ctx.content.Targets) ||
		consumedKeywords != len(ctx.content.Keywords) ||
		consumedReferences != len(ctx.content.References) ||
		consumedConditions != len(ctx.content.Conditions) ||
		len(sequence) != len(ctx.content.Effects) {
		return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — unconsumed targets/references/keywords")
	}
	return game.Mode{Targets: targets, Sequence: sequence}.Ability(), nil
}

// lowerCombinedSequenceShapes attempts the special-case combined-shape lowerers
// (single continuous effects spread across two clauses) that do not model
// per-effect conditions. It only runs when the sequence carries no conditions,
// so a condition can never be silently dropped.
func lowerCombinedSequenceShapes(cardName string, ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Conditions) != 0 {
		return game.AbilityContent{}, false
	}
	if content, ok := lowerTemporaryPTKeywordSpell(ctx); ok {
		return content, true
	}
	if content, ok := lowerGroupTemporaryPTKeywordSpell(ctx); ok {
		return content, true
	}
	if content, ok := lowerCyclingCountDamageAndGain(cardName, ctx); ok {
		return content, true
	}
	if content, ok := lowerGroupLinkedLifeSpell(ctx); ok {
		return content, true
	}
	return game.AbilityContent{}, false
}

// applyEffectConditionGate attaches an effect-gate condition to every
// instruction a gated effect produced. It returns false (fail closed) if the
// effect produced no instructions, or if any instruction already carries a
// condition, so a gate can never be silently dropped or double-applied.
func applyEffectConditionGate(sequence []game.Instruction, condition *game.EffectCondition) bool {
	if len(sequence) == 0 {
		return false
	}
	for k := range sequence {
		if sequence[k].Condition.Exists {
			return false
		}
		sequence[k].Condition = opt.Val(*condition)
	}
	return true
}

// matchSequenceEffectConditions maps each compiled condition to the single
// effect whose clause span contains it and lowers it as an effect gate. It
// returns the lowered EffectCondition keyed by effect index. ok is false (fail
// closed) if any condition is not contained in exactly one effect, if two
// conditions land on the same effect, or if a condition is not a supported
// effect-gate condition.
func matchSequenceEffectConditions(
	effects []compiler.CompiledEffect,
	conditions []compiler.CompiledCondition,
) (map[int]game.EffectCondition, bool) {
	if len(conditions) == 0 {
		return nil, true
	}
	result := make(map[int]game.EffectCondition, len(conditions))
	for ci := range conditions {
		condition := conditions[ci]
		matchIdx := -1
		for ei := range effects {
			if spanCovered(condition.Span, []shared.Span{effects[ei].Span}) {
				if matchIdx != -1 {
					return nil, false
				}
				matchIdx = ei
			}
		}
		if matchIdx == -1 {
			return nil, false
		}
		if _, exists := result[matchIdx]; exists {
			return nil, false
		}
		lowered, ok := lowerCondition(condition, conditionContextEffectGate)
		if !ok {
			return nil, false
		}
		result[matchIdx] = game.EffectCondition{
			Condition: opt.Val(lowered),
		}
	}
	return result, true
}

func localizeTargetReferences(
	references []compiler.CompiledReference,
	allTargets []compiler.CompiledTarget,
	localTargets []compiler.CompiledTarget,
) ([]compiler.CompiledReference, bool) {
	localized := append([]compiler.CompiledReference(nil), references...)
	for i := range localized {
		if localized[i].Binding != compiler.ReferenceBindingTarget {
			continue
		}
		if localized[i].Occurrence < 0 || localized[i].Occurrence >= len(allTargets) {
			return nil, false
		}
		targetSpan := allTargets[localized[i].Occurrence].Span
		local := slices.IndexFunc(localTargets, func(target compiler.CompiledTarget) bool {
			return target.Span == targetSpan
		})
		if local < 0 {
			return nil, false
		}
		localized[i].Occurrence = local
	}
	return localized, true
}

func appendReferenceAntecedentTargets(
	inherited []compiler.CompiledTarget,
	references []compiler.CompiledReference,
	allTargets []compiler.CompiledTarget,
	clauseTargets []compiler.CompiledTarget,
) []compiler.CompiledTarget {
	for _, reference := range references {
		if reference.Binding != compiler.ReferenceBindingTarget ||
			reference.Occurrence < 0 ||
			reference.Occurrence >= len(allTargets) {
			continue
		}
		target := allTargets[reference.Occurrence]
		if !oracleTargetSpanIn(target.Span, clauseTargets) &&
			!oracleTargetSpanIn(target.Span, inherited) {
			inherited = append(inherited, target)
		}
	}
	return inherited
}

func lowerDelayedTargetReturn(
	effectIndex int,
	ctx contentCtx,
	sequence []game.Instruction,
) (game.ModifyPT, game.AbilityContent, bool) {
	if effectIndex == 0 ||
		len(sequence) != effectIndex ||
		len(ctx.content.Effects) != 1 ||
		ctx.content.Effects[0].Kind != compiler.EffectReturn ||
		ctx.content.Effects[0].DelayedTiming != game.DelayedAtBeginningOfNextEndStep ||
		ctx.content.Effects[0].Negated ||
		ctx.optional ||
		ctx.content.Effects[0].Context != parser.EffectContextController ||
		ctx.content.Effects[0].ToZone != zone.Hand ||
		ctx.content.Effects[0].CounterKindKnown ||
		!referencesBindTo(ctx.content.References, compiler.ReferenceBindingTarget, 0) {
		return game.ModifyPT{}, game.AbilityContent{}, false
	}
	previous := sequence[effectIndex-1].Primitive
	if previous.Kind() != game.PrimitiveModifyPT {
		return game.ModifyPT{}, game.AbilityContent{}, false
	}
	modify, ok := previous.(game.ModifyPT)
	if !ok ||
		modify.Object.Kind() != game.ObjectReferenceTargetPermanent ||
		modify.PublishLinked != "" {
		return game.ModifyPT{}, game.AbilityContent{}, false
	}
	consumed := ctx
	consumed.content.References = nil
	consumed.content.Targets = nil
	if consumed.content.Unconsumed() {
		return game.ModifyPT{}, game.AbilityContent{}, false
	}
	key := game.LinkedKey(fmt.Sprintf("delayed-target-%d", effectIndex))
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		TargetLinkedKey: key,
	})
	if !ok {
		return game.ModifyPT{}, game.AbilityContent{}, false
	}
	modify.PublishLinked = key
	delayed := game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
		Timing: game.DelayedAtBeginningOfNextEndStep,
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.Bounce{
			Object: object,
		}}}}.Ability(),
	}}
	return modify, game.Mode{Sequence: []game.Instruction{{Primitive: delayed}}}.Ability(), true
}

// isDelayedTargetSacrificeEffect reports whether effect is a delayed
// "sacrifice it/that creature at the beginning of the next end step" clause whose
// subject is the permanent targeted by an earlier effect in the same sequence.
func isDelayedTargetSacrificeEffect(effect *compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectSacrifice &&
		effect.DelayedTiming == game.DelayedAtBeginningOfNextEndStep &&
		!effect.Negated &&
		effect.Context == parser.EffectContextController &&
		!effect.CounterKindKnown &&
		referencesBindTo(effect.References, compiler.ReferenceBindingTarget, 0)
}

// lowerDelayedSequenceClause attempts the linked delayed-trigger clause shapes
// (sacrifice, return-to-hand, and blink-return) that capture an earlier target
// and resolve it at a later step. When the clause matches one of these shapes it
// rewrites the publishing instruction in sequence and returns the delayed-trigger
// content with handled set. failed reports a matched-but-unlinkable sacrifice
// clause so the caller can fail closed. handled is false when no delayed shape
// applies and the caller should lower the clause normally.
func lowerDelayedSequenceClause(
	effects []compiler.CompiledEffect,
	effectIndex int,
	ctx contentCtx,
	sequence []game.Instruction,
) (content game.AbilityContent, handled, failed bool) {
	if isDelayedTargetSacrificeEffect(&effects[effectIndex]) {
		publisher, delayed, ok := lowerDelayedTargetSacrifice(effectIndex, ctx, sequence)
		if !ok {
			return game.AbilityContent{}, true, true
		}
		sequence[len(sequence)-1].Primitive = publisher
		return delayed, true, false
	}
	if modify, delayed, ok := lowerDelayedTargetReturn(effectIndex, ctx, sequence); ok {
		sequence[len(sequence)-1].Primitive = modify
		return delayed, true, false
	}
	if exile, delayed, ok := lowerDelayedBlinkReturn(effects, effectIndex, ctx, sequence); ok {
		sequence[len(sequence)-1].Primitive = exile
		return delayed, true, false
	}
	return game.AbilityContent{}, false, false
}

// lowerDelayedTargetSacrifice lowers a delayed "sacrifice it at the beginning of
// the next end step" clause that refers to the permanent targeted by the
// immediately preceding effect (e.g. "Target creature you control gains flying
// until end of turn. Sacrifice it at the beginning of the next end step."). The
// preceding instruction publishes the resolved target under a linked key and the
// delayed trigger sacrifices that linked object, so the captured permanent is
// sacrificed rather than the source. It returns the rewritten publishing
// primitive and the delayed-trigger content, or false to fail closed.
func lowerDelayedTargetSacrifice(
	effectIndex int,
	ctx contentCtx,
	sequence []game.Instruction,
) (game.Primitive, game.AbilityContent, bool) {
	if effectIndex == 0 ||
		len(sequence) != effectIndex ||
		len(ctx.content.Effects) != 1 ||
		!isDelayedTargetSacrificeEffect(&ctx.content.Effects[0]) ||
		ctx.optional ||
		!referencesBindTo(ctx.content.References, compiler.ReferenceBindingTarget, 0) {
		return nil, game.AbilityContent{}, false
	}
	key := game.LinkedKey(fmt.Sprintf("delayed-sacrifice-%d", effectIndex))
	publisher, ok := publishLinkedTargetPermanent(sequence[effectIndex-1].Primitive, key)
	if !ok {
		return nil, game.AbilityContent{}, false
	}
	consumed := ctx
	consumed.content.References = nil
	consumed.content.Targets = nil
	if consumed.content.Unconsumed() {
		return nil, game.AbilityContent{}, false
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		TargetLinkedKey: key,
	})
	if !ok {
		return nil, game.AbilityContent{}, false
	}
	delayed := game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
		Timing: game.DelayedAtBeginningOfNextEndStep,
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.Sacrifice{
			Object: object,
		}}}}.Ability(),
	}}
	return publisher, game.Mode{Sequence: []game.Instruction{{Primitive: delayed}}}.Ability(), true
}

// publishLinkedTargetPermanent rewrites a power/toughness or keyword-granting
// primitive that targets a permanent so it records that permanent under key for a
// later linked effect. It returns the rewritten primitive, or false when the
// primitive does not target a permanent or already publishes a linked object.
func publishLinkedTargetPermanent(primitive game.Primitive, key game.LinkedKey) (game.Primitive, bool) {
	if primitive.Kind() == game.PrimitiveModifyPT {
		modify, ok := primitive.(game.ModifyPT)
		if !ok ||
			modify.Object.Kind() != game.ObjectReferenceTargetPermanent ||
			modify.PublishLinked != "" {
			return nil, false
		}
		modify.PublishLinked = key
		return modify, true
	}
	if primitive.Kind() == game.PrimitiveApplyContinuous {
		apply, ok := primitive.(game.ApplyContinuous)
		if !ok ||
			!apply.Object.Exists ||
			apply.Object.Val.Kind() != game.ObjectReferenceTargetPermanent ||
			apply.PublishLinked != "" {
			return nil, false
		}
		apply.PublishLinked = key
		return apply, true
	}
	return nil, false
}

func lowerDelayedBlinkReturn(
	effects []compiler.CompiledEffect,
	effectIndex int,
	ctx contentCtx,
	sequence []game.Instruction,
) (game.Exile, game.AbilityContent, bool) {
	if effectIndex == 0 ||
		len(sequence) != effectIndex ||
		effects[effectIndex-1].Kind != compiler.EffectExile ||
		effects[effectIndex-1].DelayedTiming != 0 ||
		len(ctx.content.Effects) != 1 ||
		ctx.content.Effects[0].Kind != compiler.EffectReturn ||
		ctx.content.Effects[0].DelayedTiming != game.DelayedAtBeginningOfNextEndStep ||
		ctx.content.Effects[0].Negated ||
		ctx.content.Effects[0].ToZone != zone.Battlefield ||
		ctx.content.Effects[0].UnderYourControl ||
		ctx.content.Effects[0].CounterKindKnown {
		return game.Exile{}, game.AbilityContent{}, false
	}
	if !referencesBindTo(ctx.content.References, compiler.ReferenceBindingPriorInstructionResult, effectIndex-1) {
		return game.Exile{}, game.AbilityContent{}, false
	}
	// References validated — clear before fail-closed check.
	consumed := ctx
	consumed.content.References = nil
	if consumed.content.Unconsumed() {
		return game.Exile{}, game.AbilityContent{}, false
	}
	exile, ok := sequence[effectIndex-1].Primitive.(game.Exile)
	if !ok ||
		exile.Group.Valid() ||
		exile.Object.Kind() != game.ObjectReferenceTargetPermanent ||
		exile.ExileLinkedKey != "" {
		return game.Exile{}, game.AbilityContent{}, false
	}
	key := game.LinkedKey(fmt.Sprintf("delayed-blink-%d", effectIndex))
	if _, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		PriorInstruction: effectIndex - 1,
		PriorLinkedKey:   key,
	}); !ok {
		return game.Exile{}, game.AbilityContent{}, false
	}
	exile.ExileLinkedKey = key
	delayed := game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
		Timing: game.DelayedAtBeginningOfNextEndStep,
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.PutOnBattlefield{
			Source: game.LinkedBattlefieldSource(key),
		}}}}.Ability(),
	}}
	return exile, game.Mode{Sequence: []game.Instruction{{Primitive: delayed}}}.Ability(), true
}

// joinedTokenText reconstructs the source text from a token slice, inserting
// spaces between tokens where appropriate (following oracle punctuation rules).
// This mirrors the unexported compiler.joinedSourceText function.
func joinedTokenText(tokens []shared.Token) string {
	if len(tokens) == 0 {
		return ""
	}
	var b strings.Builder
	for i, tok := range tokens {
		if i > 0 && joinedTokenNeedsSpace(tokens[i-1], tok) { //nolint:gosec // i>0 guarantees valid index
			_ = b.WriteByte(' ')
		}
		_, _ = b.WriteString(tok.Text)
	}
	return b.String()
}

// upperFirst returns s with its first byte uppercased. It is safe for ASCII
// oracle text where the first character is always a plain letter.
func upperFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// sharedTargetRebaseOffset returns the accumulated-targets start index for the
// first inherited target oracle span, by looking it up in oracleSpanToGameIdx.
// The offset is used to rebase the sequence of an inherited shared-target
// clause (e.g. the "then draws" in "mills …, then draws …") so that its
// local target index 0 maps to the correct position in the already-accumulated
// targets slice, even when an earlier unrelated effect already contributed
// target specs at indices 0, 1, etc.
//
// Returns (0, false) if inherited is empty or the first span has no entry in
// the map (caller should treat this as fail-closed).
// oracleTargetSpanIn reports whether any target in list has the given span.
func oracleTargetSpanIn(span shared.Span, list []compiler.CompiledTarget) bool {
	for _, t := range list {
		if t.Span == span {
			return true
		}
	}
	return false
}

func sharedTargetRebaseOffset(inherited []compiler.CompiledTarget, spanToIdx map[shared.Span]int) (int, bool) {
	if len(inherited) == 0 {
		return 0, false
	}
	idx, ok := spanToIdx[inherited[0].Span]
	return idx, ok
}

// applyTargetRemapping sequences mode's target references to the correct
// accumulated game indices and updates the targets slice and oracleSpanToGameIdx
// map accordingly. It handles three cases:
//   - allSharedTargets: uniform rebase to the inherited target's recorded index.
//   - mixedTargets: non-uniform per-local-index remap for inherited+owned targets.
//   - default: uniform rebase starting at len(accum).
//
// Returns the updated accum slice (false if any remapping step fails).
func applyTargetRemapping(
	mode game.Mode,
	allSharedTargets, mixedTargets bool,
	inherited, owned []compiler.CompiledTarget,
	accum []game.TargetSpec,
	spanToIdx map[shared.Span]int,
) ([]game.TargetSpec, bool) {
	m := mode
	switch {
	case len(m.Targets) > 0 && allSharedTargets:
		rebaseOffset, ok := sharedTargetRebaseOffset(inherited, spanToIdx)
		if !ok || !rebaseTargetedSequence(m.Sequence, rebaseOffset) {
			return nil, false
		}
	case len(m.Targets) == 0 && allSharedTargets:
		// A shared-target clause that owns no target spec still embeds the
		// inherited antecedent's clause-local index in its primitives (e.g. a
		// CreateToken whose Recipient is the controller of the inherited target).
		// When the antecedent is not the first accumulated game target, rebase
		// that index so the reference is not silently left pointing at game target
		// 0. When the offset is zero (antecedent is the first game target),
		// existing shared clauses are left exactly as-is.
		if rebaseOffset, ok := sharedTargetRebaseOffset(inherited, spanToIdx); ok && rebaseOffset != 0 {
			if !rebaseTargetedSequence(m.Sequence, rebaseOffset) {
				return nil, false
			}
		}
	case len(m.Targets) > 0 && mixedTargets:
		if len(m.Targets) != len(inherited)+len(owned) {
			return nil, false
		}
		localToGame := make([]int, len(m.Targets))
		for j, t := range inherited {
			idx, ok := spanToIdx[t.Span]
			if !ok {
				return nil, false
			}
			localToGame[j] = idx
		}
		gameStartForOwn := len(accum)
		for j, ot := range owned {
			localToGame[len(inherited)+j] = gameStartForOwn + j
			spanToIdx[ot.Span] = gameStartForOwn + j
		}
		if !remapTargetedSequence(m.Sequence, localToGame) {
			return nil, false
		}
		accum = append(accum, m.Targets[len(inherited):]...)
	case len(m.Targets) > 0:
		gameStartIdx := len(accum)
		if !rebaseTargetedSequence(m.Sequence, gameStartIdx) {
			return nil, false
		}
		for j, ot := range owned {
			if j < len(m.Targets) {
				spanToIdx[ot.Span] = gameStartIdx + j
			}
		}
		accum = append(accum, m.Targets...)
	default:
	}
	return accum, true
}

func joinedTokenNeedsSpace(prev, cur shared.Token) bool {
	if cur.Kind == shared.Comma || cur.Kind == shared.Period || cur.Kind == shared.Colon ||
		cur.Kind == shared.Semicolon || cur.Kind == shared.RightParen ||
		cur.Kind == shared.Apostrophe || prev.Kind == shared.Apostrophe ||
		prev.Kind == shared.LeftParen || prev.Kind == shared.Quote || cur.Kind == shared.Quote {
		return false
	}
	if prev.Kind == shared.Plus || prev.Kind == shared.Minus || prev.Kind == shared.Slash ||
		cur.Kind == shared.Plus || cur.Kind == shared.Minus || cur.Kind == shared.Slash ||
		prev.Kind == shared.Asterisk || cur.Kind == shared.Asterisk {
		return false
	}
	return true
}

func lowerCyclingCountDamageAndGain(_ string, ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		ctx.content.Effects[0].Kind != compiler.EffectDealDamage ||
		ctx.content.Effects[1].Kind != compiler.EffectGain ||
		ctx.content.Effects[0].Context != parser.EffectContextSource ||
		ctx.content.Effects[1].Context != parser.EffectContextController ||
		ctx.content.Effects[1].Connection != parser.EffectConnectionAnd ||
		ctx.content.Effects[0].Negated ||
		ctx.content.Effects[1].Negated ||
		len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!singleSelfReference(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	amountEffect := ctx.content.Effects[1].Amount
	if amountEffect.DynamicKind == compiler.DynamicAmountNone ||
		amountEffect.DynamicForm != compiler.DynamicAmountWhereX ||
		!ctx.content.Effects[0].Amount.VariableX {
		return game.AbilityContent{}, false
	}
	dynamic, ok := lowerDynamicAmount(amountEffect, game.SourcePermanentReference())
	if !ok {
		return game.AbilityContent{}, false
	}
	target, ok := damageTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	amount := game.Dynamic(dynamic)
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{
			{Primitive: game.Damage{
				Amount:    amount,
				Recipient: game.AnyTargetDamageRecipient(0),
			}},
			{Primitive: game.GainLife{
				Amount: amount,
				Player: game.ControllerReference(),
			}},
		},
	}.Ability(), true
}

// lowerGroupLinkedLifeSpell handles linked two-effect patterns of the form
// "Each opponent loses N life and you gain [N | that much] life."
// It emits LoseLife with PublishResult "life-change" followed by GainLife.
// For "that much", the GainLife amount uses DynamicAmountPreviousEffectResult.
func lowerGroupLinkedLifeSpell(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		ctx.content.Effects[0].Kind != compiler.EffectLose ||
		ctx.content.Effects[1].Kind != compiler.EffectGain ||
		ctx.content.Effects[0].Context != parser.EffectContextEachOpponent ||
		ctx.content.Effects[1].Context != parser.EffectContextController ||
		ctx.content.Effects[1].Connection != parser.EffectConnectionAnd ||
		ctx.content.Effects[0].Negated ||
		ctx.content.Effects[1].Negated ||
		!ctx.content.Effects[0].Amount.Known ||
		ctx.content.Effects[0].Amount.Value < 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}
	loseAmount := game.Fixed(ctx.content.Effects[0].Amount.Value)

	// Determine the gain amount: fixed if effects[1] has a known value, dynamic ("that much") otherwise.
	var gainAmount game.Quantity
	switch {
	case ctx.content.Effects[1].Amount.Known && ctx.content.Effects[1].Amount.Value > 0:
		gainAmount = game.Fixed(ctx.content.Effects[1].Amount.Value)
	case !ctx.content.Effects[1].Amount.Known:
		gainAmount = game.Dynamic(game.DynamicAmount{
			Kind:      game.DynamicAmountPreviousEffectResult,
			ResultKey: "life-change",
		})
	default:
		return game.AbilityContent{}, false
	}

	return game.Mode{
		Sequence: []game.Instruction{
			{
				Primitive:     game.LoseLife{PlayerGroup: game.OpponentsReference(), Amount: loseAmount},
				PublishResult: "life-change",
			},
			{
				Primitive: game.GainLife{Player: game.ControllerReference(), Amount: gainAmount},
			},
		},
	}.Ability(), true
}
