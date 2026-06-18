package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerSingleOptionalEffect lowers a one-effect body whose sole effect carries
// resolving optionality ("You may draw a card.", "You may sacrifice a
// creature."). It strips the leading "you may", lowers the now-mandatory effect
// through the normal single-effect path, then marks the produced instruction
// Optional so the engine asks the controller whether to perform it (the runtime
// declines by skipping the instruction; see effectResolver.resolveInstruction).
//
// It returns ok=false (so the caller fails closed with the generic unsupported
// diagnostic) unless the body is exactly one optional effect that lowers to a
// single non-modal, no-shared-target, single-instruction sequence with no
// existing optional/result-gate/publish envelope. Anything else — an
// ability-level "you may", modes, a delayed or negated optional, or a multi-
// instruction lowering — is left unsupported rather than lowered to a
// silently-wrong sequence.
func lowerSingleOptionalEffect(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Effects) != 1 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if !effect.Optional ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.OptionalSpan.Start != effect.Span.Start ||
		effect.VerbSpan.Start.Offset <= effect.Span.Start.Offset {
		return game.AbilityContent{}, false
	}
	strippedCtx, strippedSyntax, ok := stripLeadingOptionalEffect(ctx, syntax, &effect)
	if !ok {
		return game.AbilityContent{}, false
	}
	content, diagnostic := lowerContent(cardName, strippedCtx, &strippedSyntax)
	if diagnostic != nil {
		return game.AbilityContent{}, false
	}
	if !markSingleInstructionOptional(&content) {
		return game.AbilityContent{}, false
	}
	return content, true
}

// lowerOptionalSearchSpell lowers a library-search tutor that carries resolving
// optionality on its leading search effect ("You may search your library for a
// basic land card, reveal it, put it into your hand, then shuffle."). A tutor
// compiles to several effects (search, optionally reveal, put, shuffle) that
// share one span and lower to the single game.Search instruction produced by
// lowerSearchSpell. The "you may" attaches to the search effect, so the whole
// tutor is optional: this clears the search effect's resolving optionality,
// lowers the now-mandatory tutor through lowerSearchSpell, then marks the single
// produced instruction Optional so the engine asks the controller whether to
// search.
//
// It fails closed (ok=false) unless the body is exactly one optional search
// tutor: a body-level optional, a modal body, a non-leading or non-search first
// effect, a negated/delayed search, an additional optional effect, or a tutor
// lowerSearchSpell rejects all leave the body unsupported rather than lowering a
// silently-wrong sequence.
func lowerOptionalSearchSpell(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Effects) == 0 {
		return game.AbilityContent{}, false
	}
	search := ctx.content.Effects[0]
	if search.Kind != compiler.EffectSearch ||
		!search.Optional ||
		search.Negated ||
		search.DelayedTiming != 0 {
		return game.AbilityContent{}, false
	}
	// Only the leading search effect may carry the optionality; the trailing
	// tutor effects (reveal/put/shuffle) are part of the same resolving "you
	// may" and must not independently carry optionality.
	for i := 1; i < len(ctx.content.Effects); i++ {
		if ctx.content.Effects[i].Optional {
			return game.AbilityContent{}, false
		}
	}
	stripped := ctx
	stripped.content.Effects = slices.Clone(ctx.content.Effects)
	stripped.content.Effects[0].Optional = false
	stripped.content.Effects[0].OptionalSpan = shared.Span{}
	content, diagnostic := lowerSearchSpell(stripped)
	if diagnostic != nil {
		return game.AbilityContent{}, false
	}
	if !markSingleInstructionOptional(&content) {
		return game.AbilityContent{}, false
	}
	return content, true
}

// stripLeadingOptionalEffect rebuilds ctx and syntax with the optional effect's
// "you may" prefix removed so the now-mandatory effect lowers through the
// standard single-effect path. It mirrors the optional-stripping prepareTriggerBody
// performs for trigger-level "you may". It fails closed if no body tokens remain
// after the verb.
func stripLeadingOptionalEffect(
	ctx contentCtx,
	syntax *parser.Ability,
	effect *compiler.CompiledEffect,
) (contentCtx, parser.Ability, bool) {
	verbStart := effect.VerbSpan.Start
	textOffset := verbStart.Offset - syntax.Span.Start.Offset
	if textOffset < 0 || textOffset > len(syntax.Text) {
		return contentCtx{}, parser.Ability{}, false
	}
	bodyTokens := parser.TokensFrom(syntax.Tokens, verbStart.Offset)
	if len(bodyTokens) == 0 {
		return contentCtx{}, parser.Ability{}, false
	}

	stripped := *effect
	stripped.Text = effect.Text[verbStart.Offset-effect.Span.Start.Offset:]
	stripped.Span.Start = verbStart
	stripped.Optional = false
	stripped.OptionalSpan = shared.Span{}

	// Slice the body source text (not the effect text) so syntax.Text and
	// syntax.Span stay length-aligned for offset-relative consumers such as
	// textWithoutDelimited; only the span start moves to the verb.
	strippedSyntax := *syntax
	strippedSyntax.Text = titleFirst(syntax.Text[textOffset:])
	strippedSyntax.Span.Start = verbStart
	strippedSyntax.Tokens = bodyTokens

	ctx.content.Effects = []compiler.CompiledEffect{stripped}
	ctx.span = strippedSyntax.Span
	ctx.text = strippedSyntax.Text
	return ctx, strippedSyntax, true
}

// markSingleInstructionOptional marks the sole instruction of a non-modal,
// single-mode, no-shared-target ability content Optional. Mode targets are
// permitted: the spell or ability chooses its target as normal when it is put on
// the stack, and the runtime then asks whether to apply the optional effect on
// resolution. It fails closed when the content is modal, shares targets, lowers
// to more than one instruction, or the instruction already carries an
// optional/publish/result-gate envelope — keeping the optional flow faithful to
// a single optional instruction.
func markSingleInstructionOptional(content *game.AbilityContent) bool {
	if content.IsModal() ||
		len(content.SharedTargets) != 0 ||
		len(content.Modes) != 1 ||
		len(content.Modes[0].Sequence) != 1 {
		return false
	}
	instr := &content.Modes[0].Sequence[0]
	if instr.Optional ||
		instr.PublishResult != "" ||
		instr.ResultGate.Exists {
		return false
	}
	instr.Optional = true
	return true
}

// optionalIfYouDoResultKey is the result key wiring an optional "you may X"
// instruction to its gated "if you do, Y" follow-up.
const optionalIfYouDoResultKey = game.ResultKey("if-you-do")

// optionalFlowPlan describes how an ordered effect sequence realizes resolving
// optionality. Two shapes are supported:
//
//   - "you may <X>. If you do, <Y>." (enabled): effect optionalIndex is performed
//     optionally and publishes its result, and effect gateIndex (= optionalIndex+1)
//     is gated on that result having succeeded. gateCondition is the index into the
//     content conditions of the affirmative "if you do" clause, which the sequence
//     consumes as the gate rather than as an ordinary effect condition.
//   - a trailing bare "you may <X>." (bareIndex >= 0): the final effect carries
//     resolving optionality with no "if you do" follow-up. Only that effect's own
//     instruction is marked Optional; the preceding effects are mandatory and
//     independent. bareIndex is the index of the optional effect (always the last
//     effect). enabled is false and gateIndex/gateCondition are unused in this
//     shape.
//
// bareIndex is -1 whenever the bare shape does not apply.
type optionalFlowPlan struct {
	enabled       bool
	optionalIndex int
	gateIndex     int
	gateCondition int
	bareIndex     int
}

// marksOptional reports whether the optional flow marks the instruction produced
// by effect i Optional: the optional effect of an "if you do" pair, or the
// trailing bare optional effect.
func (p optionalFlowPlan) marksOptional(i int) bool {
	return (p.enabled && i == p.optionalIndex) || i == p.bareIndex
}

// planOptionalFlow inspects an ordered effect sequence for resolving
// optionality. It returns a disabled plan (bareIndex -1) and ok=true when the
// sequence carries no resolving optionality (normal lowering proceeds
// unchanged). It returns an enabled plan for the "you may X. If you do, Y" pair,
// or a bareIndex plan for a single trailing "you may X." It returns ok=false
// (fail closed) when optionality is present but does not form one of those
// supported shapes, so the caller rejects rather than lowering a silently-wrong
// sequence.
func planOptionalFlow(content compiler.AbilityContent) (optionalFlowPlan, bool) {
	optionalIndex := -1
	for i := range content.Effects {
		if content.Effects[i].Optional {
			if optionalIndex != -1 {
				return optionalFlowPlan{}, false
			}
			optionalIndex = i
		}
	}
	if optionalIndex == -1 {
		return optionalFlowPlan{bareIndex: -1}, true
	}
	// Count "if you do" (prior-instruction-accepted) conditions. Their presence
	// selects the gated "you may X. If you do, Y" shape; their absence selects
	// the bare trailing-optional shape.
	priorAcceptedConditions := 0
	for ci := range content.Conditions {
		if content.Conditions[ci].Predicate == compiler.ConditionPredicatePriorInstructionAccepted {
			priorAcceptedConditions++
		}
	}
	if priorAcceptedConditions == 0 {
		// Bare trailing optional: the optional effect must be the final effect so
		// no later mandatory effect silently resolves as though gated on the
		// optional's result. A negated or delayed optional is left unsupported.
		if optionalIndex != len(content.Effects)-1 ||
			content.Effects[optionalIndex].Negated ||
			content.Effects[optionalIndex].DelayedTiming != 0 {
			return optionalFlowPlan{}, false
		}
		return optionalFlowPlan{optionalIndex: optionalIndex, bareIndex: optionalIndex}, true
	}
	gateIndex := optionalIndex + 1
	// The gated effect must be the final effect: any effect after it (such as an
	// "Otherwise, Z" branch, which carries no gating condition) would otherwise
	// lower as an ungated instruction and resolve unconditionally — silently
	// wrong. Restricting the flow to a tail "you may X. If you do, Y" keeps it
	// fail closed.
	if gateIndex != len(content.Effects)-1 ||
		content.Effects[gateIndex].Optional ||
		content.Effects[optionalIndex].Negated ||
		content.Effects[gateIndex].Negated ||
		content.Effects[optionalIndex].DelayedTiming != 0 ||
		content.Effects[gateIndex].DelayedTiming != 0 {
		return optionalFlowPlan{}, false
	}
	gateCondition := -1
	for ci := range content.Conditions {
		condition := content.Conditions[ci]
		if condition.Predicate != compiler.ConditionPredicatePriorInstructionAccepted {
			continue
		}
		if gateCondition != -1 ||
			condition.Kind != compiler.ConditionIf ||
			condition.Negated ||
			condition.Intervening ||
			!content.Effects[gateIndex].Order.Contains(condition.Order) ||
			content.Effects[optionalIndex].Order.Contains(condition.Order) {
			return optionalFlowPlan{}, false
		}
		gateCondition = ci
	}
	if gateCondition == -1 {
		return optionalFlowPlan{}, false
	}
	return optionalFlowPlan{
		enabled:       true,
		optionalIndex: optionalIndex,
		gateIndex:     gateIndex,
		gateCondition: gateCondition,
		bareIndex:     -1,
	}, true
}

// applyBareOptional marks the single instruction produced by a trailing bare
// "you may X" effect Optional so the engine asks the controller whether to
// perform it. It fails closed unless the effect lowered to exactly one
// instruction with no existing envelope wiring, matching lowerSingleOptionalEffect's
// single-instruction restriction.
func applyBareOptional(sequence []game.Instruction) bool {
	if len(sequence) != 1 ||
		sequence[0].Optional ||
		sequence[0].PublishResult != "" ||
		sequence[0].ResultGate.Exists {
		return false
	}
	sequence[0].Optional = true
	return true
}

// applyOptionalFlowPublish marks the single instruction produced by the optional
// effect so the engine asks the controller whether to perform it and records the
// outcome under optionalIfYouDoResultKey. It fails closed unless the optional
// effect lowered to exactly one instruction with no existing envelope wiring.
func applyOptionalFlowPublish(sequence []game.Instruction) bool {
	if len(sequence) != 1 ||
		sequence[0].Optional ||
		sequence[0].PublishResult != "" ||
		sequence[0].ResultGate.Exists {
		return false
	}
	sequence[0].Optional = true
	sequence[0].PublishResult = optionalIfYouDoResultKey
	return true
}

// applyOptionalFlowGate gates every instruction produced by the "if you do"
// effect on the optional effect having succeeded. It fails closed if any
// instruction already carries a result gate.
func applyOptionalFlowGate(sequence []game.Instruction) bool {
	if len(sequence) == 0 {
		return false
	}
	for k := range sequence {
		if sequence[k].ResultGate.Exists {
			return false
		}
		sequence[k].ResultGate = opt.Val(game.InstructionResultGate{
			Key:       optionalIfYouDoResultKey,
			Succeeded: game.TriTrue,
		})
	}
	return true
}

// optionalFlowGateConditions returns the content conditions excluding the
// affirmative "if you do" clause, which the optional flow consumes as its gate
// rather than as an ordinary per-effect condition. When the plan is disabled the
// conditions are returned unchanged.
func optionalFlowGateConditions(
	conditions []compiler.CompiledCondition,
	plan optionalFlowPlan,
) []compiler.CompiledCondition {
	if !plan.enabled {
		return conditions
	}
	filtered := make([]compiler.CompiledCondition, 0, len(conditions))
	for ci := range conditions {
		if ci == plan.gateCondition {
			continue
		}
		filtered = append(filtered, conditions[ci])
	}
	return filtered
}

// applyOptionalFlowEnvelope wires the optional-flow Optional/PublishResult and
// ResultGate onto the lowered instructions for effect i. It returns a failure
// category and false when the optionality cannot be realized, keeping the
// sequence fail closed.
func applyOptionalFlowEnvelope(plan optionalFlowPlan, i int, sequence []game.Instruction) (string, bool) {
	if plan.enabled {
		if i == plan.optionalIndex && !applyOptionalFlowPublish(sequence) {
			return "structural — optional effect not single-instruction", false
		}
		if i == plan.gateIndex && !applyOptionalFlowGate(sequence) {
			return "structural — if-you-do gate not applicable", false
		}
	}
	if i == plan.bareIndex && !applyBareOptional(sequence) {
		return "structural — optional effect not single-instruction", false
	}
	return "", true
}

// prepareSequenceClause resolves the effect at index i for per-clause lowering:
// it rebinds a prior-subject context, suppresses the optional flag for the
// optional-flow effect (its optionality is realized by the envelope instruction
// instead), and builds the clause parser.Ability with its sentence-start text
// restored. syntaxWithinSpan always clears Text, so it is restored from the
// effect text for independent effects (same span) or from the capitalised joined
// token text for then-joined sub-clauses (split span).
func prepareSequenceClause(
	ctx contentCtx,
	plan optionalFlowPlan,
	clauseSyntaxes []parser.Ability,
	i int,
) (compiler.CompiledEffect, parser.Ability) {
	effect := &ctx.content.Effects[i]
	resolvedEffect := *effect
	if effect.Context == parser.EffectContextPriorSubject {
		resolvedEffect.Context = priorSubjectContext(ctx.content.Effects, i)
	}
	if plan.marksOptional(i) {
		resolvedEffect.Optional = false
	}
	clauseAbility := clauseSyntaxes[i]
	if clauseAbility.Span != effect.Span {
		if clauseText := joinedTokenText(clauseAbility.Tokens); clauseText != "" {
			clauseAbility.Text = upperFirst(clauseText)
		}
	} else {
		clauseAbility.Text = effect.Text
	}
	return resolvedEffect, clauseAbility
}
