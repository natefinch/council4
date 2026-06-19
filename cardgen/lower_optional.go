package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const unlessPaidResultKey = game.ResultKey("unless-paid")

// lowerEventPlayerTaxedOptionalControllerBenefit lowers a targetless
// "you may <benefit> unless that player pays <mana>" trigger body. The event
// player is offered payment first; only a declined or impossible payment offers
// the resolving ability's controller the separate optional benefit.
func lowerEventPlayerTaxedOptionalControllerBenefit(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, bool) {
	if ctx.triggerEvent == game.EventUnknown ||
		ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Conditions) != 1 ||
		len(ctx.content.References) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	payment := effect.Payment
	condition := ctx.content.Conditions[0]
	reference := ctx.content.References[0]
	if !effect.Exact ||
		!effect.Optional ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Context != parser.EffectContextController ||
		effect.OptionalSpan.Start != effect.Span.Start ||
		effect.VerbSpan.Start.Offset <= effect.Span.Start.Offset ||
		payment.Payer != parser.EffectPaymentPayerEventPlayer ||
		len(payment.ManaCost) == 0 ||
		manaCostHasVariableSymbol(payment.ManaCost) ||
		condition.Kind != compiler.ConditionUnless ||
		condition.Predicate != compiler.ConditionPredicateEventPlayerDoesNotPay ||
		!condition.Order.Contains(payment.Order) ||
		reference.Kind != compiler.ReferenceThatPlayer ||
		reference.Binding != compiler.ReferenceBindingEventPlayer ||
		payment.Span.Start.Offset > reference.Span.Start.Offset ||
		payment.Span.End.Offset < reference.Span.End.Offset {
		return game.AbilityContent{}, false
	}

	strippedCtx := ctx
	strippedCtx.content.Conditions = nil
	strippedCtx.content.References = nil
	effect.Payment = compiler.CompiledEffectPayment{}
	strippedCtx.content.Effects = []compiler.CompiledEffect{effect}
	strippedCtx, strippedSyntax, ok := stripLeadingOptionalEffect(strippedCtx, syntax, &effect)
	if !ok {
		return game.AbilityContent{}, false
	}
	content, diagnostic := lowerContent(cardName, strippedCtx, &strippedSyntax)
	if diagnostic != nil ||
		content.IsModal() ||
		len(content.SharedTargets) != 0 ||
		len(content.Modes) != 1 ||
		len(content.Modes[0].Targets) != 0 ||
		len(content.Modes[0].Sequence) != 1 {
		return game.AbilityContent{}, false
	}
	benefit := content.Modes[0].Sequence[0]
	if benefit.Optional ||
		benefit.PublishResult != "" ||
		benefit.ResultGate.Exists ||
		!instructionBenefitsController(effect.Kind, benefit.Primitive) {
		return game.AbilityContent{}, false
	}
	benefit.Optional = true
	benefit.ResultGate = opt.Val(game.InstructionResultGate{
		Key:       unlessPaidResultKey,
		Succeeded: game.TriFalse,
	})
	return game.Mode{Sequence: []game.Instruction{
		{
			Primitive: game.Pay{Payment: game.ResolutionPayment{
				Prompt:   "Pay " + payment.ManaCost.String() + "?",
				Payer:    opt.Val(game.EventPlayerReference()),
				ManaCost: opt.Val(slices.Clone(payment.ManaCost)),
			}},
			PublishResult: unlessPaidResultKey,
		},
		benefit,
	}}.Ability(), true
}

func instructionBenefitsController(kind compiler.EffectKind, primitive game.Primitive) bool {
	controller := game.ControllerReference()
	switch kind {
	case compiler.EffectDraw:
		draw, ok := primitive.(game.Draw)
		return ok && draw.Player == controller && draw.PlayerGroup.Kind == game.PlayerGroupReferenceNone
	case compiler.EffectGain:
		gain, ok := primitive.(game.GainLife)
		return ok && gain.Player == controller && gain.PlayerGroup.Kind == game.PlayerGroupReferenceNone
	case compiler.EffectScry:
		scry, ok := primitive.(game.Scry)
		return ok && scry.Player == controller
	case compiler.EffectSurveil:
		surveil, ok := primitive.(game.Surveil)
		return ok && surveil.Player == controller
	case compiler.EffectInvestigate:
		investigate, ok := primitive.(game.Investigate)
		return ok && (!investigate.Recipient.Exists || investigate.Recipient.Val == controller)
	case compiler.EffectCreate:
		create, ok := primitive.(game.CreateToken)
		return ok && (!create.Recipient.Exists || create.Recipient.Val == controller)
	case compiler.EffectAddMana:
		_, ok := primitive.(game.AddMana)
		return ok
	case compiler.EffectDiscover:
		_, ok := primitive.(game.DiscoverCards)
		return ok
	case compiler.EffectProliferate:
		_, ok := primitive.(game.Proliferate)
		return ok
	default:
		return false
	}
}

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

// lowerRemovalThenControllerSearch lowers a targeted removal spell or ability
// that compensates the affected permanent's controller with an optional basic-
// land fetch — the Path to Exile / Assassin's Trophy / Cleansing Wildfire rider:
//
//	Exile target creature. Its controller may search their library for a basic
//	land card, put it onto the battlefield tapped, then shuffle.
//
// The body compiles to a mandatory leading removal effect (Exile or Destroy of a
// single target permanent) followed by an optional library-search group (search,
// optionally reveal, put, then shuffle). The tutor's grammatical subject is the
// removal *target's* controller ("Its controller"), so the search runs from that
// player's library and they — not the spell's controller — choose whether to
// search. The removal lowers through the standard single-effect path; the tutor
// lowers to one game.Search whose Player is the controller of the target (read
// from last-known information after the permanent leaves the battlefield) and
// whose instruction is Optional with the same player as OptionalActor, so the
// affected player declines by skipping the search-and-shuffle entirely.
//
// It fails closed (ok=false) unless the body is exactly this shape: a body-level
// optional, a modal body, a non-single ability target, a leading effect that is
// not single-target Exile/Destroy by the controller, a tutor whose subject is
// not the target's controller, a trailing search effect that independently
// carries optionality, or a search group searchGroupSpec cannot model all leave
// the body unsupported rather than lowered to a silently-wrong sequence.
func lowerRemovalThenControllerSearch(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Effects) < 4 {
		return game.AbilityContent{}, false
	}
	removal := ctx.content.Effects[0]
	if (removal.Kind != compiler.EffectExile && removal.Kind != compiler.EffectDestroy) ||
		removal.Optional ||
		removal.Negated ||
		removal.DelayedTiming != 0 ||
		removal.Context != parser.EffectContextController ||
		len(removal.Targets) != 1 {
		return game.AbilityContent{}, false
	}
	searchEffects := ctx.content.Effects[1:]
	search := searchEffects[0]
	if search.Kind != compiler.EffectSearch ||
		!search.Optional ||
		search.Negated ||
		search.DelayedTiming != 0 {
		return game.AbilityContent{}, false
	}
	// The tutor's subject must be the removal target's controller: exactly one
	// possessive "its" subject reference bound to the target. This is what makes
	// the affected player — not the spell's controller — the searcher and the
	// decision-maker.
	if len(search.SubjectReferences) != 1 ||
		search.SubjectReferences[0].Pronoun != compiler.ReferencePronounIts ||
		search.SubjectReferences[0].Binding != compiler.ReferenceBindingTarget {
		return game.AbilityContent{}, false
	}
	// Only the leading search effect of the tutor group may carry the "may"; the
	// trailing reveal/put/shuffle effects ride the same resolving optionality.
	for i := 1; i < len(searchEffects); i++ {
		if searchEffects[i].Optional {
			return game.AbilityContent{}, false
		}
	}
	spec, amount, ok := searchGroupSpec(searchEffects)
	if !ok {
		return game.AbilityContent{}, false
	}
	removalContent, ok := lowerRemovalClause(cardName, ctx, syntax, &removal)
	if !ok {
		return game.AbilityContent{}, false
	}
	// The removal lowered to a single non-modal, single-target instruction whose
	// target occupies ability target index 0. The tutor's controller reference
	// reads that target via TargetPermanentReference(0).
	if removalContent.IsModal() ||
		len(removalContent.Modes) != 1 ||
		len(removalContent.Modes[0].Sequence) != 1 ||
		len(removalContent.SharedTargets)+len(removalContent.Modes[0].Targets) != 1 {
		return game.AbilityContent{}, false
	}
	if removalContent.Modes[0].Sequence[0].Optional ||
		removalContent.Modes[0].Sequence[0].OptionalActor.Exists {
		return game.AbilityContent{}, false
	}
	searcher := game.ObjectControllerReference(game.TargetPermanentReference(0))
	removalContent.Modes[0].Sequence = append(removalContent.Modes[0].Sequence, game.Instruction{
		Primitive: game.Search{
			Player: searcher,
			Spec:   spec,
			Amount: game.Fixed(amount),
		},
		Optional:      true,
		OptionalActor: opt.Val(searcher),
	})
	return removalContent, true
}

// lowerRemovalClause lowers the leading removal effect of a removal-then-search
// body in isolation through the standard single-effect path, restoring its
// sentence-start clause text so offset-relative exactness consumers stay aligned.
// It fails closed if the removal does not lower cleanly on its own.
func lowerRemovalClause(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
	removal *compiler.CompiledEffect,
) (game.AbilityContent, bool) {
	removalCtx := contextForEffect(ctx, removal)
	clause := splitEffectSyntaxes(syntax, ctx.content.Effects)[0]
	clause.Text = removal.Text
	content, diagnostic := lowerContent(cardName, removalCtx, &clause)
	if diagnostic != nil {
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
//   - "you may <X>. If you do, <Y> [and <Z> ...]." (enabled): effect optionalIndex
//     is performed optionally and publishes its result, and every effect from
//     gateIndex (= optionalIndex+1) through the final effect is gated on that
//     result having succeeded. A single "if you do" clause may govern several
//     and-joined trailing effects ("If you do, draw a card and put a +1/+1
//     counter on this creature"); each compiles to its own effect and all of
//     them structurally contain the gate condition, so they form one contiguous
//     gated tail. gateCondition is the index into the content conditions of the
//     affirmative "if you do" clause, which the sequence consumes as the gate
//     rather than as an ordinary effect condition.
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

// gates reports whether the optional flow gates the instructions produced by
// effect i on the optional effect having succeeded. Every effect from gateIndex
// through the end of the sequence belongs to the "if you do" clause.
func (p optionalFlowPlan) gates(i int) bool {
	return p.enabled && i >= p.gateIndex
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
	// The optional effect must be followed by at least one gated effect and must
	// not itself be negated or delayed.
	if gateIndex >= len(content.Effects) ||
		content.Effects[optionalIndex].Negated ||
		content.Effects[optionalIndex].DelayedTiming != 0 {
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
			content.Effects[optionalIndex].Order.Contains(condition.Order) {
			return optionalFlowPlan{}, false
		}
		gateCondition = ci
	}
	if gateCondition == -1 {
		return optionalFlowPlan{}, false
	}
	// Every effect after the optional one must belong to the single "if you do"
	// clause: one affirmative "if you do" may govern several and-joined trailing
	// effects ("If you do, draw a card and put a +1/+1 counter on this
	// creature"), each compiled as its own effect that structurally contains the
	// gate condition. Requiring containment for every trailing effect rejects an
	// independent tail ("... If you do, Y. Then Z.") whose Z does not contain the
	// gate condition and would otherwise resolve unconditionally — silently
	// wrong. A negated, delayed, or independently-optional trailing effect also
	// leaves the flow unsupported.
	gateConditionOrder := content.Conditions[gateCondition].Order
	for i := gateIndex; i < len(content.Effects); i++ {
		if content.Effects[i].Optional ||
			content.Effects[i].Negated ||
			content.Effects[i].DelayedTiming != 0 ||
			!content.Effects[i].Order.Contains(gateConditionOrder) {
			return optionalFlowPlan{}, false
		}
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
		if plan.gates(i) && !applyOptionalFlowGate(sequence) {
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

// lowerOptionalHaveEffect lowers a two-effect body whose leading effect is the
// optional causative "you may have <subject> <verb> ..." ("you may have this
// creature deal 1 damage to target player", "you may have target opponent
// discard a card"). The parser models the causative "have"/"has" as a leading
// EffectGrantKeyword carrying the resolving optionality, with the real action
// (deal damage, discard, ...) compiled as a second effect sharing the same
// sentence span. The "have" effect carries no keyword payload of its own — it is
// purely structural — so this drops it, lowers the real action effect as a
// single mandatory instruction through the normal single-effect path, then marks
// that instruction Optional.
//
// It fails closed (ok=false) unless the body is exactly this controller "you may
// have <subject> <action>" shape lowering to one non-modal, no-shared-target,
// single-instruction sequence: a body-level optional, a modal body, a
// non-controller "<player> may have" (whose causative "have" is not the ability
// controller), a negated or delayed action, or an action the single-effect path
// cannot lower all leave the body unsupported rather than lowered to a
// silently-wrong sequence.
func lowerOptionalHaveEffect(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Effects) != 2 ||
		len(ctx.content.Keywords) != 0 {
		return game.AbilityContent{}, false
	}
	have := ctx.content.Effects[0]
	action := ctx.content.Effects[1]
	// The causative "have"/"has" compiles to an EffectGrantKeyword that grants no
	// keyword of its own: it is purely structural, so the ability content carries
	// no compiled keyword (checked above) and the real action rides as a second
	// effect. Requiring the "have" to belong to the controller
	// (EffectContextController is the compiled form of the "you may have ..."
	// subject) rejects the non-controller "<player> may have <subject> <action>"
	// shape (for example "that creature's controller may have it deal ..."),
	// which the runtime cannot model as a controller-gated optional. Requiring
	// the verb to start after the sentence start rejects a sentence-leading grant.
	if have.Kind != compiler.EffectGrantKeyword ||
		!have.Optional ||
		have.Context != parser.EffectContextController ||
		have.Negated ||
		have.DelayedTiming != 0 ||
		have.OptionalSpan.Start != have.Span.Start ||
		have.VerbSpan.Start.Offset <= have.Span.Start.Offset {
		return game.AbilityContent{}, false
	}
	// The action must be the same-sentence consequence of the causative "have".
	// Requiring identical spans rejects any independent trailing effect. The
	// action may itself carry optionality: a controller "you may have it deal ..."
	// (pronoun/referenced-object subject) marks both effects optional, so its
	// optionality is cleared during stripping below rather than rejected here.
	if action.Span != have.Span ||
		action.Negated ||
		action.DelayedTiming != 0 {
		return game.AbilityContent{}, false
	}
	strippedCtx := ctx
	strippedAction := action
	strippedAction.Optional = false
	strippedAction.OptionalSpan = shared.Span{}
	// The action carried RequiresOrderedLowering only because the ability had a
	// second effect (the structural "have"); as the now-sole effect it lowers
	// through the standard single-effect path.
	strippedAction.RequiresOrderedLowering = false
	strippedCtx.content.Effects = []compiler.CompiledEffect{strippedAction}
	content, diagnostic := lowerContent(cardName, strippedCtx, syntax)
	if diagnostic != nil {
		return game.AbilityContent{}, false
	}
	if !markSingleInstructionOptional(&content) {
		return game.AbilityContent{}, false
	}
	return content, true
}

// lowerOptionalBlinkReturn lowers the optional immediate-blink (flicker) body —
// "You may exile [another] target <permanent>, then return that card to the
// battlefield under [its owner's / your] control." — the Conjurer's Closet /
// Soulherder / Felidar Guardian / Wispweaver Angel shape. The "you may" attaches
// to the leading exile effect, and the trailing ", then return that card" clause
// back-references the exiled card. The whole flow is optional at resolution: the
// controller chooses the target when the spell or ability goes on the stack, then
// decides on resolution whether to exile-and-return.
//
// The body compiles to two effects sharing the blink semantics: a leading
// single-target Exile carrying the resolving optionality and a trailing immediate
// Return whose object binds to the exile's result. This clears the exile's
// optionality, lowers the now-mandatory blink through the ordered effect-sequence
// path (which produces the two-instruction [Exile, PutOnBattlefield] sequence
// lowerImmediateBlinkReturn builds, with the exile rewritten to remember the
// exiled card under a linked key), then marks the exile instruction Optional and
// publishing and gates the put on the exile having succeeded. Declining the exile
// publishes a not-accepted result, so the gated put is skipped and nothing
// returns; accepting it exiles the target and returns it, exactly honoring the
// controller's choice on both branches.
//
// It fails closed (ok=false) unless the body is exactly this controller
// optional-exile-then-immediate-return shape lowering to one non-modal,
// no-shared-target [Exile, PutOnBattlefield] sequence: a body-level optional, a
// modal body, a non-single ability target, a non-optional or negated/delayed
// exile, a non-controller exile, an independently-optional return, a delayed
// ("at the beginning of the next end step") return, or any lowering that does not
// produce the exact two-instruction blink sequence all leave the body unsupported
// rather than lowered to a silently-wrong sequence.
func lowerOptionalBlinkReturn(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 1 {
		return game.AbilityContent{}, false
	}
	exile := ctx.content.Effects[0]
	ret := ctx.content.Effects[1]
	// The leading effect must be the controller's optional single-target exile,
	// with the "you may" attached at the effect's sentence start so the whole
	// blink — not some interior clause — is the optional action.
	if exile.Kind != compiler.EffectExile ||
		!exile.Optional ||
		exile.Negated ||
		exile.DelayedTiming != 0 ||
		exile.Context != parser.EffectContextController ||
		exile.OptionalSpan.Start != exile.Span.Start ||
		exile.VerbSpan.Start.Offset <= exile.Span.Start.Offset ||
		len(exile.Targets) != 1 {
		return game.AbilityContent{}, false
	}
	// The trailing return must be the immediate ", then return that card to the
	// battlefield" blink form and must not carry independent optionality: its
	// optionality rides the same resolving "you may" as the exile. A delayed
	// return ("at the beginning of the next end step") is left unsupported here so
	// only the same-resolution blink is gated on the exile result.
	if ret.Kind != compiler.EffectReturn ||
		ret.Optional ||
		ret.Negated ||
		ret.Connection != parser.EffectConnectionThen ||
		ret.DelayedTiming != 0 ||
		ret.ToZone != zone.Battlefield {
		return game.AbilityContent{}, false
	}
	// Clear the exile's resolving optionality and lower the now-mandatory blink
	// through the ordered effect-sequence path, which links the exile to the
	// return and validates that every target and reference is consumed.
	stripped := ctx
	stripped.content.Effects = slices.Clone(ctx.content.Effects)
	stripped.content.Effects[0].Optional = false
	stripped.content.Effects[0].OptionalSpan = shared.Span{}
	content, diagnostic := lowerOrderedEffectSequence(cardName, stripped, syntax)
	if diagnostic != nil {
		return game.AbilityContent{}, false
	}
	if !markBlinkExileOptional(&content) {
		return game.AbilityContent{}, false
	}
	return content, true
}

// markBlinkExileOptional marks the leading Exile instruction of a lowered
// immediate-blink sequence Optional and publishing under optionalIfYouDoResultKey,
// and gates the trailing PutOnBattlefield instruction on that exile having
// succeeded. It fails closed unless the content is a single non-modal,
// no-shared-target mode whose sequence is exactly [Exile, PutOnBattlefield] with
// no existing optional/publish/result-gate envelope — keeping the optional flow
// faithful to the two-instruction blink shape.
func markBlinkExileOptional(content *game.AbilityContent) bool {
	if content.IsModal() ||
		len(content.SharedTargets) != 0 ||
		len(content.Modes) != 1 ||
		len(content.Modes[0].Sequence) != 2 {
		return false
	}
	exile := &content.Modes[0].Sequence[0]
	put := &content.Modes[0].Sequence[1]
	if exile.Primitive == nil ||
		exile.Primitive.Kind() != game.PrimitiveExile ||
		exile.Optional ||
		exile.PublishResult != "" ||
		exile.ResultGate.Exists ||
		exile.OptionalActor.Exists {
		return false
	}
	if put.Primitive == nil ||
		put.Primitive.Kind() != game.PrimitivePutOnBattlefield ||
		put.Optional ||
		put.PublishResult != "" ||
		put.ResultGate.Exists {
		return false
	}
	exile.Optional = true
	exile.PublishResult = optionalIfYouDoResultKey
	put.ResultGate = opt.Val(game.InstructionResultGate{
		Key:       optionalIfYouDoResultKey,
		Succeeded: game.TriTrue,
	})
	return true
}
