package cardgen

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const unlessPaidResultKey = game.ResultKey("unless-paid")
const controllerPaidResultKey = game.ResultKey("controller-paid")

// lowerControllerPaidEffect lowers a "you may pay <cost>. If you do, <body>."
// resolution whose consequence begins with a controller verb (destroy, draw, put
// a counter, ...). The optional payment becomes a resolution Pay instruction
// publishing its result, and every consequence instruction is gated on that
// payment having succeeded. A consequence that targets a single object (such as
// "destroy target creature") is supported: its mode target is promoted onto the
// resulting ability mode, so the target is chosen when the ability goes on the
// stack and the gated effect applies to it only if the controller pays.
func lowerControllerPaidEffect(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) == 0 ||
		len(ctx.content.Conditions) != 1 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	// The folded payment marks one consequence effect (mana costs leave no
	// prelude effect; non-mana costs such as "sacrifice a land" leave their own
	// cost effect ahead of the consequence). Locate that payment-bearing effect;
	// any effects before it are the cost prelude already captured by the payment
	// and are dropped, while effects after it complete the consequence body.
	payIdx := -1
	for i := range ctx.content.Effects {
		if ctx.content.Effects[i].Payment.Form == parser.EffectPaymentFormMayPayThenIfDo {
			if payIdx != -1 {
				return game.AbilityContent{}, false
			}
			payIdx = i
		}
	}
	if payIdx == -1 {
		return game.AbilityContent{}, false
	}
	for i := 0; i < payIdx; i++ {
		if !ctx.content.Effects[i].Exact ||
			ctx.content.Effects[i].Context != parser.EffectContextController {
			return game.AbilityContent{}, false
		}
	}
	effect := ctx.content.Effects[payIdx]
	restEffects := ctx.content.Effects[payIdx+1:]
	payment := effect.Payment
	condition := ctx.content.Conditions[0]
	hasMana := len(payment.ManaCost) != 0
	hasAdditional := payment.AdditionalCost != nil
	if effect.Context != parser.EffectContextController ||
		!effect.Exact ||
		effect.Optional ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		payment.Form != parser.EffectPaymentFormMayPayThenIfDo ||
		payment.Payer != parser.EffectPaymentPayerController ||
		(!hasMana && !hasAdditional) ||
		(hasMana && manaCostHasVariableSymbol(payment.ManaCost)) ||
		payment.GenericManaAmount.DynamicKind != compiler.DynamicAmountNone ||
		condition.Kind != compiler.ConditionIf ||
		condition.Predicate != compiler.ConditionPredicatePriorInstructionAccepted ||
		condition.NodeID != payment.SuccessConditionNodeID ||
		payment.Span.End.Offset >= condition.Span.Start.Offset {
		return game.AbilityContent{}, false
	}
	resolutionPayment, ok := controllerPaidResolutionPayment(cardName, payment)
	if !ok {
		return game.AbilityContent{}, false
	}

	strippedCtx := ctx
	strippedCtx.content.Conditions = nil
	effect.Payment = compiler.CompiledEffectPayment{}
	strippedCtx.content.Effects = []compiler.CompiledEffect{effect}
	strippedCtx, strippedSyntax, ok := stripEffectPrefix(strippedCtx, syntax, &effect)
	if !ok {
		return game.AbilityContent{}, false
	}
	// stripEffectPrefix keeps only the payment-bearing first effect and shifts its
	// span to the consequence verb; restore the rest of the consequence
	// sentence's effects (such as the "put them onto the battlefield tapped, then
	// shuffle" tail of a search) so the consequence lowers as a complete
	// standalone body. The search lowering requires every merged effect to share
	// one span, so align the trailing effects with the stripped first effect.
	alignedSpan := strippedCtx.content.Effects[0].Span
	for i := range restEffects {
		restEffects[i].Span = alignedSpan
	}
	strippedCtx.content.Effects = append(strippedCtx.content.Effects, restEffects...)
	// The dropped prefix cost effects (a non-mana payment such as "discard a
	// card" parses to its own effect in the payment sentence) inflate the
	// ability's legacy-effect count, which marks every sibling effect
	// RequiresOrderedLowering. Now that those cost effects are captured by the
	// resolution payment and removed, the isolated consequence stands alone, so
	// clear the flag: a single consequence effect lowers directly and a
	// multi-effect consequence still routes through ordered lowering by count.
	for i := range strippedCtx.content.Effects {
		strippedCtx.content.Effects[i].RequiresOrderedLowering = false
	}
	content, diagnostic := lowerContent(cardName, strippedCtx, &strippedSyntax)
	if diagnostic != nil ||
		content.IsModal() ||
		len(content.SharedTargets) != 0 ||
		len(content.Modes) != 1 ||
		len(content.Modes[0].Sequence) == 0 {
		return game.AbilityContent{}, false
	}
	consequenceTargets := content.Modes[0].Targets
	consequence := content.Modes[0].Sequence
	for i := range consequence {
		if consequence[i].Optional ||
			consequence[i].PublishResult != "" ||
			consequence[i].ResultGate.Exists {
			return game.AbilityContent{}, false
		}
		consequence[i].ResultGate = opt.Val(game.InstructionResultGate{
			Key:       controllerPaidResultKey,
			Succeeded: game.TriTrue,
		})
	}
	sequence := make([]game.Instruction, 0, len(consequence)+1)
	sequence = append(sequence, game.Instruction{
		Primitive:     game.Pay{Payment: resolutionPayment},
		PublishResult: controllerPaidResultKey,
	})
	sequence = append(sequence, consequence...)
	return game.Mode{Targets: consequenceTargets, Sequence: sequence}.Ability(), true
}

// lowerOptionalPaidBenefit lowers a "you may pay {mana}. If you do, <body>."
// resolution whose consequence body begins with a non-controller-context effect,
// such as the Extort drain "each opponent loses 1 life and you gain that much
// life." Unlike lowerControllerPaidEffect it does not strip the consequence to a
// single controller verb; it lowers the entire consequence body through the
// shared content path and then gates every resulting instruction on the optional
// payment. The controller, verb-initial family is handled by
// lowerControllerPaidEffect, so this path keys on a non-controller leading
// effect to keep the two disjoint. The consequence may be a single targeted
// rider ("target player loses 1 life", "this enchantment deals 1 damage to any
// target") or a multi-effect body ("target player loses 1 life and you gain that
// much life"). When the consequence targets a single object, its mode target is
// promoted onto the resulting ability mode, so the target is chosen when the
// ability goes on the stack and the gated benefit applies only if the controller
// pays.
func lowerOptionalPaidBenefit(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) == 0 ||
		len(ctx.content.Conditions) != 1 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	payIdx := -1
	for i := range ctx.content.Effects {
		if ctx.content.Effects[i].Payment.Form == parser.EffectPaymentFormMayPayThenIfDo {
			if payIdx != -1 {
				return game.AbilityContent{}, false
			}
			payIdx = i
		}
	}
	if payIdx != 0 || ctx.content.Effects[0].Context == parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	payment := ctx.content.Effects[0].Payment
	condition := ctx.content.Conditions[0]
	if len(payment.ManaCost) == 0 ||
		payment.AdditionalCost != nil ||
		payment.Form != parser.EffectPaymentFormMayPayThenIfDo ||
		payment.Payer != parser.EffectPaymentPayerController ||
		manaCostHasVariableSymbol(payment.ManaCost) ||
		payment.GenericManaAmount.DynamicKind != compiler.DynamicAmountNone ||
		condition.Kind != compiler.ConditionIf ||
		condition.Predicate != compiler.ConditionPredicatePriorInstructionAccepted ||
		condition.NodeID != payment.SuccessConditionNodeID ||
		payment.Span.End.Offset >= condition.Span.Start.Offset {
		return game.AbilityContent{}, false
	}
	resolutionPayment, ok := controllerPaidResolutionPayment(cardName, payment)
	if !ok {
		return game.AbilityContent{}, false
	}

	bodyCtx := ctx
	bodyCtx.content.Conditions = nil
	bodyEffects := slices.Clone(ctx.content.Effects)
	bodyEffects[0].Payment = compiler.CompiledEffectPayment{}
	bodyCtx.content.Effects = bodyEffects
	content, diagnostic := lowerContent(cardName, bodyCtx, syntax)
	if diagnostic != nil ||
		content.IsModal() ||
		len(content.SharedTargets) != 0 ||
		len(content.Modes) != 1 ||
		len(content.Modes[0].Sequence) == 0 {
		return game.AbilityContent{}, false
	}
	consequenceTargets := content.Modes[0].Targets
	consequence := content.Modes[0].Sequence
	for i := range consequence {
		if consequence[i].Optional ||
			consequence[i].ResultGate.Exists {
			return game.AbilityContent{}, false
		}
		consequence[i].ResultGate = opt.Val(game.InstructionResultGate{
			Key:       controllerPaidResultKey,
			Succeeded: game.TriTrue,
		})
	}
	sequence := make([]game.Instruction, 0, len(consequence)+1)
	sequence = append(sequence, game.Instruction{
		Primitive:     game.Pay{Payment: resolutionPayment},
		PublishResult: controllerPaidResultKey,
	})
	sequence = append(sequence, consequence...)
	return game.Mode{Targets: consequenceTargets, Sequence: sequence}.Ability(), true
}

// controllerPaidResolutionPayment builds the runtime resolution payment for a
// "you may <cost>. If you do, ..." controller payment from its compiled mana or
// additional cost. It fails closed when an additional cost carries a mana
// component or no lowerable additional cost.
func controllerPaidResolutionPayment(cardName string, payment compiler.CompiledEffectPayment) (game.ResolutionPayment, bool) {
	hasMana := len(payment.ManaCost) != 0
	hasAdditional := payment.AdditionalCost != nil
	if !hasMana && !hasAdditional {
		return game.ResolutionPayment{}, false
	}
	resolution := game.ResolutionPayment{}
	var prompts []string
	if hasMana {
		resolution.ManaCost = opt.Val(slices.Clone(payment.ManaCost))
		prompts = append(prompts, "Pay "+payment.ManaCost.String())
	}
	if hasAdditional {
		manaCost, additionalCosts, ok := lowerActivationCostComponents(cardName, payment.AdditionalCost)
		if !ok || manaCost != nil || len(additionalCosts) == 0 {
			return game.ResolutionPayment{}, false
		}
		resolution.AdditionalCosts = additionalCosts
		for i := range additionalCosts {
			prompts = append(prompts, additionalCosts[i].Text)
		}
	}
	resolution.Prompt = upperFirst(strings.Join(prompts, " and ")) + "?"
	return resolution, true
}

// lowerEventPlayerTaxedControllerBenefit lowers a targetless event-player mana
// payment followed by a controller benefit. It supports both "you may <benefit>
// unless that player pays" and "that player may pay. If the player doesn't,
// <benefit>", preserving whether the benefit itself is optional.
func lowerEventPlayerTaxedControllerBenefit(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, bool) {
	if ctx.triggerEvent == game.EventUnknown ||
		ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Conditions) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	payment := effect.Payment
	condition := ctx.content.Conditions[0]
	eventPlayerReference, referencesOK := eventPlayerPaymentReferences(ctx.content.References, payment)
	if !effect.Exact ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Context != parser.EffectContextController ||
		payment.Payer != parser.EffectPaymentPayerEventPlayer ||
		payment.GenericManaAmount.DynamicKind != compiler.DynamicAmountNone && effect.Kind != compiler.EffectDraw ||
		!eventPlayerPaymentCostSupported(payment) ||
		condition.Predicate != compiler.ConditionPredicateEventPlayerDoesNotPay ||
		!referencesOK ||
		payment.Span.Start.Offset > eventPlayerReference.Span.Start.Offset ||
		payment.Span.End.Offset < eventPlayerReference.Span.End.Offset {
		return game.AbilityContent{}, false
	}
	benefitOptional := false
	switch payment.Form {
	case parser.EffectPaymentFormUnless:
		if condition.Kind != compiler.ConditionUnless ||
			!condition.Order.Contains(payment.Order) &&
				(payment.GenericManaAmount.DynamicKind == compiler.DynamicAmountNone ||
					condition.Order.Start != payment.Order.Start) {
			return game.AbilityContent{}, false
		}
		if effect.Optional {
			if effect.OptionalSpan.Start != effect.Span.Start ||
				effect.VerbSpan.Start.Offset <= effect.Span.Start.Offset {
				return game.AbilityContent{}, false
			}
			benefitOptional = true
		}
	case parser.EffectPaymentFormMayPayThenIfDoesNot:
		if effect.Optional ||
			condition.Kind != compiler.ConditionIf ||
			condition.NodeID != payment.FailureConditionNodeID ||
			payment.Span.End.Offset >= condition.Span.Start.Offset {
			return game.AbilityContent{}, false
		}
	default:
		return game.AbilityContent{}, false
	}

	strippedCtx := ctx
	strippedCtx.content.Conditions = nil
	strippedCtx.content.References = nil
	effect.Payment = compiler.CompiledEffectPayment{}
	strippedCtx.content.Effects = []compiler.CompiledEffect{effect}
	strippedCtx, strippedSyntax, ok := stripEffectPrefix(strippedCtx, syntax, &effect)
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
	benefit.Optional = benefitOptional
	benefit.ResultGate = opt.Val(game.InstructionResultGate{
		Key:       unlessPaidResultKey,
		Succeeded: game.TriFalse,
	})
	resolutionPayment, ok := lowerEventPlayerResolutionPayment(payment)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{
		{
			Primitive:     game.Pay{Payment: resolutionPayment},
			PublishResult: unlessPaidResultKey,
		},
		benefit,
	}}.Ability(), true
}

const eventPlayerPaidResultKey = game.ResultKey("event-player-paid")

// lowerEventPlayerPaidBenefit lowers the affirmative event-player optional
// payment "that player may pay {N}. If the player does, <benefit>." (Unifying
// Theory's "they draw a card"), the resolving-success mirror of
// lowerEventPlayerTaxedControllerBenefit's "If the player doesn't" failure gate.
// The event player is offered the payment; the consequence resolves only when
// the player pays. The optional payment becomes a resolution Pay instruction
// charged to the event player and publishing its result, and every consequence
// instruction is gated on that payment having succeeded (TriTrue).
//
// The consequence is lowered compositionally through the shared content path
// after stripping the payment and the "if the player does" gate, so it fails
// closed on any benefit the backend cannot already lower. The payment offer's
// "that player" payer reference sits inside the payment span; it is dropped
// before the consequence lowers so a recipient reference in the benefit (the
// event-player "they" of "they draw a card") is the only reference the
// consequence sees.
func lowerEventPlayerPaidBenefit(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) == 0 ||
		len(ctx.content.Conditions) != 1 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Targets) != 0 {
		return game.AbilityContent{}, false
	}
	payIdx := -1
	for i := range ctx.content.Effects {
		if ctx.content.Effects[i].Payment.Form == parser.EffectPaymentFormMayPayThenIfDo {
			if payIdx != -1 {
				return game.AbilityContent{}, false
			}
			payIdx = i
		}
	}
	if payIdx != 0 {
		return game.AbilityContent{}, false
	}
	payment := ctx.content.Effects[0].Payment
	condition := ctx.content.Conditions[0]
	if len(payment.ManaCost) == 0 ||
		payment.AdditionalCost != nil ||
		payment.Form != parser.EffectPaymentFormMayPayThenIfDo ||
		payment.Payer != parser.EffectPaymentPayerEventPlayer ||
		manaCostHasVariableSymbol(payment.ManaCost) ||
		payment.GenericManaAmount.DynamicKind != compiler.DynamicAmountNone ||
		condition.Kind != compiler.ConditionIf ||
		condition.Predicate != compiler.ConditionPredicatePriorInstructionAccepted ||
		condition.NodeID != payment.SuccessConditionNodeID ||
		payment.Span.End.Offset >= condition.Span.Start.Offset {
		return game.AbilityContent{}, false
	}
	// Drop the payer reference ("that player") that lives inside the payment
	// span; keep every other reference (the benefit's recipient pronoun) for the
	// consequence lowering. Any reference straddling the payment boundary is
	// unexpected and fails closed.
	bodyReferences := make([]compiler.CompiledReference, 0, len(ctx.content.References))
	for _, reference := range ctx.content.References {
		if reference.Span.End.Offset <= payment.Span.End.Offset {
			if reference.Span.Start.Offset < payment.Span.Start.Offset {
				return game.AbilityContent{}, false
			}
			continue
		}
		if reference.Span.Start.Offset < payment.Span.End.Offset {
			return game.AbilityContent{}, false
		}
		bodyReferences = append(bodyReferences, reference)
	}
	resolutionPayment, ok := lowerEventPlayerResolutionPayment(payment)
	if !ok {
		return game.AbilityContent{}, false
	}

	bodyCtx := ctx
	bodyCtx.content.Conditions = nil
	bodyCtx.content.References = bodyReferences
	bodyEffects := slices.Clone(ctx.content.Effects)
	bodyEffects[0].Payment = compiler.CompiledEffectPayment{}
	bodyCtx.content.Effects = bodyEffects
	content, diagnostic := lowerContent(cardName, bodyCtx, syntax)
	if diagnostic != nil ||
		content.IsModal() ||
		len(content.SharedTargets) != 0 ||
		len(content.Modes) != 1 ||
		len(content.Modes[0].Targets) != 0 ||
		len(content.Modes[0].Sequence) == 0 {
		return game.AbilityContent{}, false
	}
	consequence := content.Modes[0].Sequence
	for i := range consequence {
		if consequence[i].Optional ||
			consequence[i].PublishResult != "" ||
			consequence[i].ResultGate.Exists {
			return game.AbilityContent{}, false
		}
		consequence[i].ResultGate = opt.Val(game.InstructionResultGate{
			Key:       eventPlayerPaidResultKey,
			Succeeded: game.TriTrue,
		})
	}
	sequence := make([]game.Instruction, 0, len(consequence)+1)
	sequence = append(sequence, game.Instruction{
		Primitive:     game.Pay{Payment: resolutionPayment},
		PublishResult: eventPlayerPaidResultKey,
	})
	sequence = append(sequence, consequence...)
	return game.Mode{Sequence: sequence}.Ability(), true
}

func eventPlayerPaymentCostSupported(payment compiler.CompiledEffectPayment) bool {
	if payment.GenericManaAmount.DynamicKind == compiler.DynamicAmountNone {
		return !manaCostHasVariableSymbol(payment.ManaCost)
	}
	return len(payment.ManaCost) == 1 &&
		payment.ManaCost[0] == cost.X &&
		payment.GenericManaAmount.DynamicKind == compiler.DynamicAmountSourcePower &&
		payment.GenericManaAmount.DynamicForm == compiler.DynamicAmountWhereX &&
		payment.GenericManaAmount.Multiplier == 1
}

func eventPlayerPaymentReferences(references []compiler.CompiledReference, payment compiler.CompiledEffectPayment) (compiler.CompiledReference, bool) {
	var eventPlayer compiler.CompiledReference
	sourcePower := false
	for _, reference := range references {
		switch {
		case reference.Kind == compiler.ReferenceThatPlayer &&
			reference.Binding == compiler.ReferenceBindingEventPlayer:
			if eventPlayer.Kind != compiler.ReferenceUnknown {
				return compiler.CompiledReference{}, false
			}
			eventPlayer = reference
		case reference.Kind == compiler.ReferencePronoun &&
			reference.Pronoun == compiler.ReferencePronounTheir &&
			reference.Binding == compiler.ReferenceBindingEventPlayer &&
			reference.Span.End.Offset <= payment.Span.Start.Offset:
		case payment.GenericManaAmount.DynamicKind == compiler.DynamicAmountSourcePower &&
			reference.Binding == compiler.ReferenceBindingSource &&
			reference.Span == payment.GenericManaAmount.ReferenceSpan:
			if sourcePower {
				return compiler.CompiledReference{}, false
			}
			sourcePower = true
		default:
			return compiler.CompiledReference{}, false
		}
	}
	wantSourcePower := payment.GenericManaAmount.DynamicKind == compiler.DynamicAmountSourcePower
	return eventPlayer, eventPlayer.Kind != compiler.ReferenceUnknown && sourcePower == wantSourcePower
}

func lowerEventPlayerResolutionPayment(payment compiler.CompiledEffectPayment) (game.ResolutionPayment, bool) {
	result := game.ResolutionPayment{
		Prompt: "Pay " + payment.ManaCost.String() + "?",
		Payer:  opt.Val(game.EventPlayerReference()),
	}
	if payment.GenericManaAmount.DynamicKind == compiler.DynamicAmountNone {
		result.ManaCost = opt.Val(slices.Clone(payment.ManaCost))
		return result, true
	}
	dynamic, ok := lowerDynamicAmount(payment.GenericManaAmount, game.SourcePermanentReference())
	if !ok || dynamic.Kind != game.DynamicAmountObjectPower {
		return game.ResolutionPayment{}, false
	}
	result.DynamicGenericManaCost = opt.Val(&dynamic)
	return result, true
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
	case compiler.EffectPopulate:
		create, ok := primitive.(game.CreateToken)
		return ok && (!create.Recipient.Exists || create.Recipient.Val == controller)
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
	strippedCtx, strippedSyntax, ok := stripEffectPrefix(ctx, syntax, &effect)
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

// lowerOptionalReferencedControllerSearch lowers a standalone optional library
// fetch performed by the controller of a referenced object — the death-trigger
// reanimation rider "When enchanted creature dies, that creature's controller may
// search their library for a creature card, put that card onto the battlefield,
// then shuffle." (Pattern of Rebirth). The search effect's grammatical subject is
// the triggering permanent's controller, so the runtime resolves the searcher to
// ObjectControllerReference(EventPermanentReference()) and that player — not the
// ability's controller — decides whether to search and chooses the card.
//
// It fails closed unless the body is exactly one optional search group whose
// subject is the event permanent's controller: a body-level optional, a modal or
// targeted body, a non-leading or non-search first effect, an independently
// optional trailing effect, or a group searchGroupSpec cannot model exactly all
// leave the body unsupported rather than lowering a silently-wrong sequence.
func lowerOptionalReferencedControllerSearch(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Effects) == 0 {
		return game.AbilityContent{}, false
	}
	search := ctx.content.Effects[0]
	if search.Kind != compiler.EffectSearch ||
		!search.Optional ||
		search.Negated ||
		search.DelayedTiming != 0 ||
		search.Context != parser.EffectContextReferencedObjectController ||
		len(search.SubjectReferences) != 1 ||
		search.SubjectReferences[0].Binding != compiler.ReferenceBindingEventPermanent {
		return game.AbilityContent{}, false
	}
	// Only the leading search effect carries the "may"; the trailing reveal/put/
	// shuffle effects ride the same resolving optionality.
	for i := 1; i < len(ctx.content.Effects); i++ {
		if ctx.content.Effects[i].Optional {
			return game.AbilityContent{}, false
		}
	}
	group, ok := searchGroupSpec(ctx.content.Effects)
	if !ok || group.Length != len(ctx.content.Effects) {
		return game.AbilityContent{}, false
	}
	searcher := game.ObjectControllerReference(game.EventPermanentReference())
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.Search{
			Player: searcher,
			Spec:   group.Spec,
			Amount: group.Quantity,
		},
		Optional:      true,
		OptionalActor: opt.Val(searcher),
	}}}.Ability(), true
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
	// The tutor's subject must be the removal target's controller: the possessive
	// "its controller" / "that <permanent>'s controller" form or "that player",
	// bound to the target.
	if len(search.SubjectReferences) != 1 {
		return game.AbilityContent{}, false
	}
	subject := search.SubjectReferences[0]
	if subject.Binding != compiler.ReferenceBindingTarget ||
		(subject.Kind != compiler.ReferenceThatPlayer &&
			subject.Kind != compiler.ReferenceThatObject &&
			(subject.Kind != compiler.ReferencePronoun ||
				subject.Pronoun != compiler.ReferencePronounIts)) {
		return game.AbilityContent{}, false
	}
	// The body may carry one fetch group (Path to Exile) or two (Demolition
	// Field: the affected player's fetch followed by the controller's own "You
	// may search ..."). Split the trailing effects at the second search so each
	// group's "then shuffle" is the final effect of its own slice, as
	// searchGroupSpec requires.
	firstLen := len(searchEffects)
	for i := 1; i < len(searchEffects); i++ {
		if searchEffects[i].Kind == compiler.EffectSearch {
			firstLen = i
			break
		}
	}
	firstEffects := searchEffects[:firstLen]
	group, ok := searchGroupSpec(firstEffects)
	if !ok || group.Length != firstLen {
		return game.AbilityContent{}, false
	}
	// Only the leading search effect of the tutor group may carry the "may"; the
	// trailing reveal/put/shuffle effects ride the same resolving optionality.
	for i := 1; i < firstLen; i++ {
		if firstEffects[i].Optional {
			return game.AbilityContent{}, false
		}
	}
	// A second optional "You may search your library ..." group may follow the
	// affected player's fetch — the controller's own basic-land fetch on land
	// destruction (Demolition Field). It is the controller's search, declined
	// independently of the first. Any other trailing shape fails closed.
	var youGroup searchGroup
	hasYouSearch := false
	if remaining := searchEffects[firstLen:]; len(remaining) > 0 {
		youSearch := remaining[0]
		if youSearch.Kind != compiler.EffectSearch ||
			!youSearch.Optional ||
			youSearch.Negated ||
			youSearch.DelayedTiming != 0 ||
			youSearch.Context != parser.EffectContextController {
			return game.AbilityContent{}, false
		}
		youGroup, ok = searchGroupSpec(remaining)
		if !ok || youGroup.Length != len(remaining) {
			return game.AbilityContent{}, false
		}
		for i := 1; i < youGroup.Length; i++ {
			if remaining[i].Optional {
				return game.AbilityContent{}, false
			}
		}
		hasYouSearch = true
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
			Spec:   group.Spec,
			Amount: group.Quantity,
		},
		Optional:      true,
		OptionalActor: opt.Val(searcher),
	})
	if hasYouSearch {
		removalContent.Modes[0].Sequence = append(removalContent.Modes[0].Sequence, game.Instruction{
			Primitive: game.Search{
				Player: game.ControllerReference(),
				Spec:   youGroup.Spec,
				Amount: youGroup.Quantity,
			},
			Optional: true,
		})
	}
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

// stripEffectPrefix rebuilds ctx and syntax from the effect verb so a recognized
// optional or conditional prefix is removed before the effect routes through the
// standard single-effect path. It fails closed if no body tokens remain.
func stripEffectPrefix(
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
//
// publishWithoutOptional selects a third, mandatory shape: "X. If you do, Y."
// where the leading effect at optionalIndex is not optional but can fail to do
// anything ("exile it. If you do, create a token"). The leading effect publishes
// whether it succeeded (without being made Optional) and the gated tail is gated
// on that success, exactly like the enabled optional shape.
//
// elseIndex marks the start of a trailing else branch in the enabled shape:
// "you may X. If you do, Y. Otherwise, Z." or the "If you don't, Z." wording. The
// gated "if you do" tail runs gateIndex..elseIndex-1; the else tail runs
// elseIndex..end. Each else effect is gated on the optional result having failed
// (the controller declined X), the exact complement of the affirmative "if you
// do" gate on Y. elseIndex is -1 whenever there is no else branch.
//
// elseGateCondition is the content-condition index of the "if you don't"
// complement gate (a ConditionPredicatePriorInstructionNotAccepted clause) when
// the else branch uses that wording; the sequence consumes it as the else gate
// rather than as an ordinary per-effect condition. It is -1 for the "Otherwise,"
// wording (whose else effect carries no condition) and whenever there is no else
// branch.
type optionalFlowPlan struct {
	enabled                bool
	optionalIndex          int
	gateIndex              int
	gateCondition          int
	bareIndex              int
	elseIndex              int
	elseGateCondition      int
	publishWithoutOptional bool
	// extraOptionalIndex is the index of a second optional effect nested inside
	// the affirmative gated tail ("you may X. If you do, you may Y"): Y is gated
	// on X having succeeded and is itself performed only if the controller then
	// chooses to. It is -1 whenever no nested optional is present.
	extraOptionalIndex int
	// independentOptional selects the all-independent bare-optional shape: a
	// gate-free sequence in which every effect carries its own resolving "you
	// may" with no relationship between them ("you may tap or untap target
	// permanent, then you may tap or untap another target permanent"). Each
	// effect lowers to a single instruction marked Optional independently, so the
	// controller decides each one separately. It is mutually exclusive with the
	// enabled/bareIndex shapes.
	independentOptional bool
	// gateOnFailure inverts the affirmative gate so the gated tail runs when the
	// publishing effect FAILED rather than succeeded ("if no life is lost this
	// way, convert" — Blitzwing, Cruel Tormentor). It applies Succeeded: TriFalse
	// to the gated instructions instead of the default TriTrue. It is meaningful
	// only for the mandatory no-life-lost flow; every other flow leaves it false.
	gateOnFailure bool
}

// gateSucceeded reports the TriState the affirmative gate requires of the
// publishing effect's result: TriFalse for the inverted no-life-lost flow,
// TriTrue for every ordinary "if you do" flow.
func (p optionalFlowPlan) gateSucceeded() game.TriState {
	if p.gateOnFailure {
		return game.TriFalse
	}
	return game.TriTrue
}

// marksOptional reports whether the optional flow marks the instruction produced
// by effect i Optional: the optional effect of an "if you do" pair, the trailing
// bare optional effect, or — in the all-independent shape — every effect.
func (p optionalFlowPlan) marksOptional(i int) bool {
	if p.independentOptional {
		return true
	}
	if p.enabled && p.extraOptionalIndex >= 0 && i == p.extraOptionalIndex {
		return true
	}
	if p.publishWithoutOptional {
		return false
	}
	return (p.enabled && i == p.optionalIndex) || i == p.bareIndex
}

// gates reports whether the optional flow gates the instructions produced by
// effect i on the optional effect having succeeded. Every effect from gateIndex
// through the end of the "if you do" clause belongs to it; the trailing
// "Otherwise" else tail (elseIndex onward) is excluded — those are gated on the
// optional effect having failed by gatesElse instead.
func (p optionalFlowPlan) gates(i int) bool {
	return p.enabled && i >= p.gateIndex && (p.elseIndex < 0 || i < p.elseIndex)
}

// gatesElse reports whether the optional flow gates the instructions produced by
// effect i on the optional effect having failed: the trailing else branch
// ("Otherwise, <Z>." or "If you don't, <Z>.") resolves only when the controller
// declined the optional effect.
func (p optionalFlowPlan) gatesElse(i int) bool {
	return p.enabled && p.elseIndex >= 0 && i >= p.elseIndex
}

// clearsNegated reports whether the optional flow should drop a parser-set
// Negated flag on effect i before lowering. The "If you don't, <Z>." wording
// leaves the word "don't" in the else effect's clause, which the effect
// classifier reads as a negation of the else action itself; that negation
// belongs to the complement gate (recorded as elseGateCondition), not to Z, so
// it is cleared on the gate-carrying else effect.
func (p optionalFlowPlan) clearsNegated(i int) bool {
	return p.enabled && p.elseGateCondition >= 0 && i == p.elseIndex
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
	// The primary optional is the first "you may X" effect; it is the one a
	// following "if/when you do" gate keys on. A nested "you may X. If you do,
	// you may Y" body carries a second optional in the gated tail (Y), which the
	// gated path below admits and lowers as its own optional, gated instruction.
	// Any optional outside the affirmative gated tail (a second independent
	// optional, or one in an else branch) is not modeled and fails closed.
	optionalIndex := -1
	extraOptional := -1
	for i := range content.Effects {
		if !content.Effects[i].Optional {
			continue
		}
		switch {
		case optionalIndex == -1:
			optionalIndex = i
		case extraOptional == -1:
			extraOptional = i
		default:
			return optionalFlowPlan{}, false
		}
	}
	if optionalIndex == -1 {
		if plan, ok, handled := planMandatoryIfYouDoFlow(content); handled {
			return plan, ok
		}
		if plan, ok, handled := planMandatoryNoLifeLostFlow(content); handled {
			return plan, ok
		}
		return optionalFlowPlan{bareIndex: -1, elseIndex: -1, elseGateCondition: -1, extraOptionalIndex: -1}, true
	}
	// Count "if you do" (prior-instruction-accepted) conditions, including the
	// outcome-worded "a <type> is destroyed this way" gate that the optional-
	// destroy shape (Noxious Gearhulk) treats identically. Their presence selects
	// the gated "you may X. If you do, Y" shape; their absence selects the bare
	// trailing-optional shape.
	priorAcceptedConditions := 0
	for ci := range content.Conditions {
		if isResolvingSuccessGate(content.Conditions[ci].Predicate) {
			priorAcceptedConditions++
		}
	}
	if priorAcceptedConditions == 0 {
		// Multiple independent bare optionals with no resolving-success gate:
		// "you may X, then you may Y[, then you may Z]." Every effect must carry
		// its own resolving "you may" so none silently resolves as though gated
		// on another's result; a leading or trailing mandatory effect, a negated
		// optional, or a delayed optional leaves the body unsupported. Each effect
		// lowers to a single instruction marked Optional independently.
		if extraOptional != -1 {
			for ei := range content.Effects {
				if !content.Effects[ei].Optional ||
					content.Effects[ei].Negated ||
					content.Effects[ei].DelayedTiming != 0 {
					return optionalFlowPlan{}, false
				}
			}
			return optionalFlowPlan{bareIndex: -1, elseIndex: -1, elseGateCondition: -1, extraOptionalIndex: -1, independentOptional: true}, true
		}
		// Bare trailing optional: the optional effect must be the final effect so
		// no later mandatory effect silently resolves as though gated on the
		// optional's result. A negated or delayed optional is left unsupported.
		if optionalIndex != len(content.Effects)-1 ||
			content.Effects[optionalIndex].Negated ||
			content.Effects[optionalIndex].DelayedTiming != 0 {
			return optionalFlowPlan{}, false
		}
		return optionalFlowPlan{optionalIndex: optionalIndex, bareIndex: optionalIndex, elseIndex: -1, elseGateCondition: -1, extraOptionalIndex: -1}, true
	}
	gateIndex := optionalIndex + 1
	// When the single optional effect is itself the gated consequence of a
	// preceding mandatory effect ("put a +1/+1 counter on it. When you do, you
	// may remove a counter from target permanent"), the optional is not the gate
	// source: a mandatory effect publishes the gate and the optional rides in the
	// gated tail. The mandatory-publish flow models that shape, marking the gated
	// optional Y both gated and Optional. Only a lone optional qualifies; a body
	// that already carries a leading optional gate source ("you may X. If you do,
	// you may Y") keeps the source-optional interpretation below.
	if extraOptional == -1 && optionalEffectIsGatedConsequence(content, optionalIndex) {
		if plan, ok, handled := planMandatoryIfYouDoFlow(content); handled {
			return plan, ok
		}
		return optionalFlowPlan{}, false
	}
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
		if !isResolvingSuccessGate(condition.Predicate) {
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
	// Every effect after the optional one belongs to either the single "if you
	// do" clause or a trailing else branch. The affirmative clause may govern
	// several and-joined trailing effects ("If you do, draw a card and put a
	// +1/+1 counter on this creature"), each compiled as its own effect that
	// structurally contains the gate condition. The else branch resolves only
	// when the controller declined the optional effect, so it does NOT contain
	// the affirmative gate condition and is gated on the optional result having
	// failed instead. It takes one of two wordings:
	//   - "Otherwise, <Z>." — a sentence-initial EffectConnectionOtherwise effect
	//     carrying no condition.
	//   - "If you don't, <Z>." — an effect carrying a complement gate condition
	//     (ConditionPredicatePriorInstructionNotAccepted), consumed as the else
	//     gate (elseGateCondition) like the affirmative "if you do" gate.
	// Requiring affirmative-gate containment for every non-else trailing effect
	// rejects an independent tail ("... If you do, Y. Then Z.") whose Z does not
	// contain the gate condition and would otherwise resolve unconditionally —
	// silently wrong. A negated or independently-optional trailing effect leaves
	// the flow unsupported. A delayed affirmative consequence ("If you do,
	// convert Ultra Magnus at end of combat.") is allowed: its
	// CreateDelayedTrigger is gated on the optional effect's result, so the
	// delayed trigger is scheduled only when the optional effect succeeded.
	gateConditionOrder := content.Conditions[gateCondition].Order
	elseIndex := -1
	elseGateCondition := -1
	for i := gateIndex; i < len(content.Effects); i++ {
		if content.Effects[i].Connection == parser.EffectConnectionOtherwise {
			elseIndex = i
			break
		}
		if ci := elseGateConditionIndex(content, i); ci >= 0 {
			elseIndex = i
			elseGateCondition = ci
			break
		}
	}
	tailEnd := len(content.Effects)
	if elseIndex >= 0 {
		tailEnd = elseIndex
	}
	for i := gateIndex; i < tailEnd; i++ {
		effect := content.Effects[i]
		if effect.Negated {
			return optionalFlowPlan{}, false
		}
		// An optional gated effect realizes the nested "you may X. If you do, you
		// may Y" body: Y is performed optionally and gated on X having succeeded.
		// Only the single recognized second optional may appear, and it must be
		// the gate-anchored consequence (it structurally contains the gate
		// condition), never a free-floating optional that would resolve outside
		// the gate.
		if effect.Optional {
			if i != extraOptional || !effect.Order.Contains(gateConditionOrder) {
				return optionalFlowPlan{}, false
			}
			continue
		}
		if effect.Order.Contains(gateConditionOrder) {
			continue
		}
		// A trailing effect that does not itself contain the gate condition is
		// allowed only when it is a prior-subject continuation of an earlier
		// gated effect ("... put a +1/+1 counter on target attacking creature.
		// It gains trample until end of turn."). Such a continuation depends on a
		// subject introduced inside the gated tail, so it belongs to the gated
		// consequence even though its own sentence does not repeat the "If/When
		// you do" gate. The first gated effect must still anchor the gate, and an
		// independent new-subject tail ("... If you do, Y. Then Z.") does not
		// qualify and stays rejected as silently-wrong.
		if i == gateIndex || !isGatedSubjectContinuation(effect) {
			return optionalFlowPlan{}, false
		}
	}
	// A recognized second optional must have been consumed inside the affirmative
	// gated tail above. One outside that range (in the else branch or before the
	// gate) has no modeled relationship to the primary optional and fails closed.
	if extraOptional != -1 && (extraOptional < gateIndex || extraOptional >= tailEnd) {
		return optionalFlowPlan{}, false
	}
	if elseIndex >= 0 {
		// The else branch requires at least one preceding gated "if you do"
		// effect, and its own effects must be plain (not optional/delayed) and
		// must not belong to the "if you do" clause (they are the opposite
		// branch, so they never contain the affirmative gate condition). The
		// "If you don't" wording leaves the word "don't" in the gate-carrying
		// else effect, which the classifier reads as a negation of Z; that
		// negation belongs to the complement gate and is cleared before lowering,
		// so it is tolerated only on that one effect.
		if elseIndex == gateIndex {
			return optionalFlowPlan{}, false
		}
		for i := elseIndex; i < len(content.Effects); i++ {
			negatedArtifact := i == elseIndex && elseGateCondition >= 0
			if content.Effects[i].Optional ||
				(content.Effects[i].Negated && !negatedArtifact) ||
				content.Effects[i].DelayedTiming != 0 ||
				content.Effects[i].Order.Contains(gateConditionOrder) {
				return optionalFlowPlan{}, false
			}
		}
	}
	return optionalFlowPlan{
		enabled:            true,
		optionalIndex:      optionalIndex,
		gateIndex:          gateIndex,
		gateCondition:      gateCondition,
		bareIndex:          -1,
		elseIndex:          elseIndex,
		elseGateCondition:  elseGateCondition,
		extraOptionalIndex: extraOptional,
	}, true
}

// optionalEffectIsGatedConsequence reports whether the optional effect at
// optionalIndex is the gated consequence of a preceding "if/when you do" gate
// rather than the gate source: its clause order contains a resolving-success
// gate condition. A source optional ("you may X. If you do, Y") never contains
// the gate (the gate lives on the following Y), so this distinguishes the
// mandatory-publish shape ("X. When you do, you may Y") from the source-optional
// shape.
func optionalEffectIsGatedConsequence(content compiler.AbilityContent, optionalIndex int) bool {
	for ci := range content.Conditions {
		condition := content.Conditions[ci]
		if !isResolvingSuccessGate(condition.Predicate) {
			continue
		}
		if content.Effects[optionalIndex].Order.Contains(condition.Order) {
			return true
		}
	}
	return false
}

// isGatedSubjectContinuation reports whether an effect grammatically continues a
// prior effect's subject ("It gains trample until end of turn." after "... put a
// +1/+1 counter on target attacking creature."). Such an effect names no subject
// of its own; it back-references the most recent subject, so when it trails a
// gated effect it belongs to the same gated consequence. Only the prior-subject
// and referenced-object continuation contexts qualify; every independent subject
// context is excluded so an unrelated trailing effect is never silently gated.
func isGatedSubjectContinuation(effect compiler.CompiledEffect) bool {
	switch effect.Context {
	case parser.EffectContextReferencedObject, parser.EffectContextPriorSubject:
		return true
	default:
		return false
	}
}

// elseGateConditionIndex returns the content-condition index of the "If you
// don't" complement gate (ConditionPredicatePriorInstructionNotAccepted)
// contained in effect i's clause order, or -1 when effect i carries no such
// well-formed gate. The gate must be a single non-intervening, non-negated "if"
// condition, mirroring the affirmative "if you do" gate; any other shape returns
// -1 so the caller leaves the flow unsupported.
func elseGateConditionIndex(content compiler.AbilityContent, effectIndex int) int {
	found := -1
	for ci := range content.Conditions {
		condition := content.Conditions[ci]
		if condition.Predicate != compiler.ConditionPredicatePriorInstructionNotAccepted {
			continue
		}
		if !content.Effects[effectIndex].Order.Contains(condition.Order) {
			continue
		}
		if found != -1 ||
			condition.Kind != compiler.ConditionIf ||
			condition.Negated ||
			condition.Intervening {
			return -1
		}
		found = ci
	}
	return found
}

// isResolvingSuccessGate reports whether a condition predicate is the affirmative
// "if you do" resolving-success gate, either the literal "if you do" or the
// outcome-worded "a <type> is destroyed this way". Both gate trailing effects on
// the immediately preceding effect having succeeded; the destroyed-this-way form
// is accepted only in the optional-destroy shape where the destroyed type matches
// the destroy target, and the mandatory result flow accepts only the literal form.
func isResolvingSuccessGate(predicate compiler.ConditionPredicate) bool {
	return predicate == compiler.ConditionPredicatePriorInstructionAccepted ||
		predicate == compiler.ConditionPredicateDestroyedThisWay
}

// planMandatoryIfYouDoFlow detects a mandatory "X. If you do, Y." result gate,
// where the leading effect is not optional but can fail to do anything ("exile
// it. If you do, create a token" — the exile is mandatory yet does nothing if
// the source already left the battlefield). The leading effect publishes whether
// it succeeded and the trailing "if you do" effects are gated on that success,
// exactly like the optional "you may X. If you do, Y." pair (CR 608.2c).
//
// handled is false when the sequence carries no "if you do" gate, so the caller
// proceeds with normal ungated lowering. When an "if you do" gate is present but
// does not form a supported mandatory pair, it returns handled=true with ok=false
// so the caller fails closed rather than dropping the gate.
func planMandatoryIfYouDoFlow(content compiler.AbilityContent) (plan optionalFlowPlan, ok bool, handled bool) {
	gateCondition := -1
	for ci := range content.Conditions {
		condition := content.Conditions[ci]
		if condition.Predicate != compiler.ConditionPredicatePriorInstructionAccepted {
			continue
		}
		if gateCondition != -1 ||
			condition.Kind != compiler.ConditionIf ||
			condition.Negated ||
			condition.Intervening {
			return optionalFlowPlan{}, false, true
		}
		gateCondition = ci
	}
	if gateCondition == -1 {
		return optionalFlowPlan{}, false, false
	}
	gateConditionOrder := content.Conditions[gateCondition].Order
	// The gated "if you do" effects are the contiguous tail whose clause Order
	// contains the gate condition; the publishing effect is the one immediately
	// before that tail.
	gateIndex := -1
	for i := range content.Effects {
		if content.Effects[i].Order.Contains(gateConditionOrder) {
			gateIndex = i
			break
		}
	}
	if gateIndex <= 0 {
		return optionalFlowPlan{}, false, true
	}
	publishIndex := gateIndex - 1
	// The publishing effect must be a plain mandatory effect that does not itself
	// belong to the gated clause.
	if content.Effects[publishIndex].Optional ||
		content.Effects[publishIndex].Negated ||
		content.Effects[publishIndex].DelayedTiming != 0 ||
		content.Effects[publishIndex].Order.Contains(gateConditionOrder) {
		return optionalFlowPlan{}, false, true
	}
	// Every effect from the gate index onward must belong to the single "if you
	// do" clause, mirroring planOptionalFlow's contiguous-tail requirement so an
	// independent ungated effect cannot silently resolve as though gated. A
	// single nested optional effect realizes the "X. When you do, you may Y"
	// reflexive body: Y is gated on X having happened and performed only if the
	// controller then chooses to. Any further optional, or one that does not
	// itself anchor the gate, fails closed.
	extraOptional := -1
	for i := gateIndex; i < len(content.Effects); i++ {
		effect := content.Effects[i]
		if effect.Negated ||
			effect.DelayedTiming != 0 ||
			!effect.Order.Contains(gateConditionOrder) {
			return optionalFlowPlan{}, false, true
		}
		if effect.Optional {
			if extraOptional != -1 {
				return optionalFlowPlan{}, false, true
			}
			extraOptional = i
		}
	}
	return optionalFlowPlan{
		enabled:                true,
		optionalIndex:          publishIndex,
		gateIndex:              gateIndex,
		gateCondition:          gateCondition,
		bareIndex:              -1,
		elseIndex:              -1,
		elseGateCondition:      -1,
		publishWithoutOptional: true,
		extraOptionalIndex:     extraOptional,
	}, true, true
}

// planMandatoryNoLifeLostFlow detects the mandatory negated resolving-success
// gate "X loses life ... If no life is lost this way, Y." (Blitzwing, Cruel
// Tormentor). The leading lose-life effect publishes whether it caused any life
// loss and the trailing gated effect resolves only when it did not — the failure
// complement of the "if you do" gate handled by planMandatoryIfYouDoFlow, wired
// as an inverted affirmative gate (gateOnFailure) so the tail is gated on
// Succeeded: TriFalse.
//
// handled is false when the sequence carries no "no life is lost this way" gate,
// so the caller proceeds with normal ungated lowering. When the gate is present
// but does not form a supported pair (its publisher is not a plain mandatory
// lose-life effect immediately before a single gated tail), it returns
// handled=true with ok=false so the caller fails closed.
func planMandatoryNoLifeLostFlow(content compiler.AbilityContent) (plan optionalFlowPlan, ok bool, handled bool) {
	gateCondition := -1
	for ci := range content.Conditions {
		condition := content.Conditions[ci]
		if condition.Predicate != compiler.ConditionPredicateNoLifeLostThisWay {
			continue
		}
		if gateCondition != -1 ||
			condition.Kind != compiler.ConditionIf ||
			condition.Negated ||
			condition.Intervening {
			return optionalFlowPlan{}, false, true
		}
		gateCondition = ci
	}
	if gateCondition == -1 {
		return optionalFlowPlan{}, false, false
	}
	gateConditionOrder := content.Conditions[gateCondition].Order
	gateIndex := -1
	for i := range content.Effects {
		if content.Effects[i].Order.Contains(gateConditionOrder) {
			gateIndex = i
			break
		}
	}
	if gateIndex <= 0 {
		return optionalFlowPlan{}, false, true
	}
	publishIndex := gateIndex - 1
	// The publishing effect must be a plain mandatory lose-life effect: "no life
	// is lost this way" reports the resolving success of a preceding life loss,
	// so any other publisher shape is not the printed antecedent and fails closed.
	if content.Effects[publishIndex].Kind != compiler.EffectLose ||
		content.Effects[publishIndex].Optional ||
		content.Effects[publishIndex].Negated ||
		content.Effects[publishIndex].DelayedTiming != 0 ||
		content.Effects[publishIndex].Order.Contains(gateConditionOrder) {
		return optionalFlowPlan{}, false, true
	}
	// Every effect from the gate index onward must belong to the single gated
	// clause and be a plain mandatory effect, mirroring planMandatoryIfYouDoFlow's
	// contiguous-tail requirement so an independent ungated effect cannot silently
	// resolve as though gated.
	for i := gateIndex; i < len(content.Effects); i++ {
		effect := content.Effects[i]
		if effect.Optional ||
			effect.Negated ||
			effect.DelayedTiming != 0 ||
			!effect.Order.Contains(gateConditionOrder) {
			return optionalFlowPlan{}, false, true
		}
	}
	return optionalFlowPlan{
		enabled:                true,
		optionalIndex:          publishIndex,
		gateIndex:              gateIndex,
		gateCondition:          gateCondition,
		bareIndex:              -1,
		elseIndex:              -1,
		elseGateCondition:      -1,
		publishWithoutOptional: true,
		extraOptionalIndex:     -1,
		gateOnFailure:          true,
	}, true, true
}

// you may Y"): it is gated on X having succeeded (ResultGate TriTrue) and is
// itself Optional so the engine asks the controller whether to perform Y only
// when X happened. The runtime evaluates the result gate before the optional
// prompt, so a declined or failed X skips the prompt entirely. It fails closed
// unless Y lowered to exactly one instruction with no existing envelope wiring.
func applyGatedOptionalFlow(sequence []game.Instruction) bool {
	if len(sequence) != 1 ||
		sequence[0].Optional ||
		sequence[0].PublishResult != "" ||
		sequence[0].ResultGate.Exists {
		return false
	}
	sequence[0].Optional = true
	sequence[0].ResultGate = opt.Val(game.InstructionResultGate{
		Key:       optionalIfYouDoResultKey,
		Succeeded: game.TriTrue,
	})
	return true
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

// applyResultFlowPublish publishes the single instruction produced by the leading
// effect of a mandatory "X. If you do, Y." pair under optionalIfYouDoResultKey
// without marking it Optional. It fails closed unless the effect lowered to
// exactly one instruction with no existing envelope wiring.
func applyResultFlowPublish(sequence []game.Instruction) bool {
	if len(sequence) != 1 ||
		sequence[0].Optional ||
		sequence[0].PublishResult != "" ||
		sequence[0].ResultGate.Exists {
		return false
	}
	sequence[0].PublishResult = optionalIfYouDoResultKey
	return true
}

// applyOptionalFlowGate gates every instruction produced by a branch effect on
// the optional effect's published result matching succeeded: TriTrue for the
// affirmative "if you do" clause, TriFalse for the "Otherwise" else branch. It
// fails closed if any instruction already carries a result gate.
func applyOptionalFlowGate(sequence []game.Instruction, succeeded game.TriState) bool {
	if len(sequence) == 0 {
		return false
	}
	for k := range sequence {
		if sequence[k].ResultGate.Exists {
			return false
		}
		sequence[k].ResultGate = opt.Val(game.InstructionResultGate{
			Key:       optionalIfYouDoResultKey,
			Succeeded: succeeded,
		})
	}
	return true
}

// optionalFlowGateConditions returns the content conditions excluding the
// affirmative "if you do" clause and, when present, the "if you don't" complement
// clause, both of which the optional flow consumes as its gates rather than as
// ordinary per-effect conditions. When the plan is disabled the conditions are
// returned unchanged.
func optionalFlowGateConditions(
	conditions []compiler.CompiledCondition,
	plan optionalFlowPlan,
) []compiler.CompiledCondition {
	if !plan.enabled {
		return conditions
	}
	filtered := make([]compiler.CompiledCondition, 0, len(conditions))
	for ci := range conditions {
		if ci == plan.gateCondition || ci == plan.elseGateCondition {
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
		if i == plan.optionalIndex {
			if plan.publishWithoutOptional {
				if !applyResultFlowPublish(sequence) {
					return "structural — result effect not single-instruction", false
				}
			} else if !applyOptionalFlowPublish(sequence) {
				return "structural — optional effect not single-instruction", false
			}
		}
		if plan.gates(i) {
			if plan.extraOptionalIndex >= 0 && i == plan.extraOptionalIndex {
				if !applyGatedOptionalFlow(sequence) {
					return "structural — gated optional not single-instruction", false
				}
			} else if !applyOptionalFlowGate(sequence, plan.gateSucceeded()) {
				return "structural — if-you-do gate not applicable", false
			}
		}
		if plan.gatesElse(i) && !applyOptionalFlowGate(sequence, game.TriFalse) {
			return "structural — otherwise gate not applicable", false
		}
	}
	if plan.independentOptional {
		if !applyBareOptional(sequence) {
			return "structural — optional effect not single-instruction", false
		}
		return "", true
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
	if plan.clearsNegated(i) {
		resolvedEffect.Negated = false
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
		len(ctx.content.Effects) != 2 {
		return game.AbilityContent{}, false
	}
	have := ctx.content.Effects[0]
	action := ctx.content.Effects[1]
	// The structural "have"/"has" grants no keyword of its own, so any compiled
	// keyword belongs to the causative action's clause ("you may have <subject>
	// gain flying until end of turn"). Requiring every keyword to lie within the
	// action's clause span keeps a stray keyword from some other clause from
	// being silently carried into the action's lowering; a keyword outside that
	// span leaves the body unsupported rather than lowered to a wrong shape.
	if len(keywordsWithinSpan(ctx.content.Keywords, action.ClauseSpan)) != len(ctx.content.Keywords) {
		return game.AbilityContent{}, false
	}
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
	if !strippedAction.Exact && causativeActionForcibleExact(&strippedAction, ctx.content.Keywords) {
		// The causative action's clause uses the base-verb form ("each player
		// draw a card", "target player mill two cards") because it is governed by
		// "have". The parser's exact reconstruction only matches the finite-verb
		// standalone form ("Each player draws a card"), so the action stays
		// non-exact even though every field the runtime needs — kind, fixed
		// amount, subject context, and any target — is fully parsed. Like the
		// RequiresOrderedLowering flag cleared above, this non-exactness is an
		// artifact of the structural "have" sibling, not a genuinely unrecognized
		// clause (HasUnrecognizedSibling is required to be false). Clear it and
		// scope the body to the action's own bindings so the subject-sensitive
		// single-effect lowerer sees exactly the standalone action it would lower
		// for the finite form, then validates the shape from those parsed fields.
		strippedAction.Exact = true
		strippedCtx.content.Targets = action.Targets
		strippedCtx.content.References = action.References
		strippedCtx.content.Keywords = keywordsWithinSpan(ctx.content.Keywords, action.ClauseSpan)
	}
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

// causativeActionForcibleExact reports whether a non-exact causative "have"
// action may have its exactness artifact cleared (see lowerOptionalHaveEffect).
// It is restricted to action kinds whose runtime effect is fully determined by
// the parser's structured fields (effect kind, amount or power/toughness deltas,
// granted keyword, subject context, and target) so that clearing the base-verb-
// only non-exactness cannot admit an unhandled clause: the corresponding single-
// effect lowerer re-validates every one of those fields and fails closed
// otherwise. A genuinely unrecognized sibling (HasUnrecognizedSibling) is never
// forcible.
//
// Each admitted kind requires the magnitude-bearing field the runtime needs to
// be fully parsed: a known fixed amount or a recognized dynamic amount for the
// count effects, known power and toughness deltas (or a recognized dynamic
// amount) for the "+X/+X / -X/-X" change, and either a known amount ("gain N
// life") or a keyword within the action clause ("gain flying until end of turn")
// for the grant, and a known fixed or recognized dynamic amount ("deal damage
// equal to its power") for the damage. Dynamic amounts are admitted rather than excluded because the
// single-effect lowerer re-validates the dynamic form and fails closed for any it
// cannot model, exactly as it does for the standalone finite-verb clause; a
// genuinely unrecognized amount still leaves the body unsupported.
func causativeActionForcibleExact(
	action *compiler.CompiledEffect,
	keywords []compiler.CompiledKeyword,
) bool {
	if action.HasUnrecognizedSibling {
		return false
	}
	// A count amount is fully parsed when it is a known fixed value or a
	// recognized dynamic quantity ("equal to the number of ..."); the downstream
	// lowerer validates the dynamic form and fails closed otherwise.
	countAmountParsed := action.Amount.Known ||
		action.Amount.DynamicKind != compiler.DynamicAmountNone
	switch action.Kind {
	case compiler.EffectDraw,
		compiler.EffectMill,
		compiler.EffectDiscard,
		compiler.EffectLose,
		compiler.EffectDealDamage,
		compiler.EffectCreate:
		// The count effects carry their magnitude in Amount. A causative "deal N
		// damage to <target>" likewise carries a known fixed value, and "deal
		// damage equal to its power" a recognized dynamic amount; a causative
		// "create N <token>" carries its token count in Amount. The downstream
		// single-effect lowerer re-validates the damage source/recipient/target
		// or the token identity (a predefined named token, or a fixed
		// power/toughness token with its types, one subtype, and at most one
		// color) and fails closed for any it cannot model.
		return countAmountParsed
	case compiler.EffectGain:
		// "gain N life" carries its magnitude in Amount; "gain <keyword> until
		// end of turn" carries the granted keyword within the action clause and
		// leaves Amount unset.
		return countAmountParsed ||
			len(keywordsWithinSpan(keywords, action.ClauseSpan)) > 0
	case compiler.EffectModifyPT:
		// A causative "get +X/+X / -X/-X until end of turn" change carries its
		// magnitude in the power/toughness deltas (fixed) or in a recognized
		// dynamic amount, never in Amount: a positive fixed pump populates Amount
		// as an artifact, but the shrink and dynamic forms do not.
		return (action.PowerDelta.Known && action.ToughnessDelta.Known) ||
			action.Amount.DynamicKind != compiler.DynamicAmountNone
	default:
		return false
	}
}

// lowerOptionalBlinkReturn lowers the optional blink (flicker) body — "You may
// exile [another] target <permanent>, then return that card to the battlefield
// under [its owner's / your] control." (immediate, Conjurer's Closet / Soulherder
// / Felidar Guardian / Wispweaver Angel) and "You may exile target <permanent>.
// If you do, return that card to the battlefield … at the beginning of the next
// end step." (delayed, Sentinel of the Pearl Trident / Astral Slide / Astral
// Drift). The "you may" attaches to the leading exile effect, and the trailing
// return clause back-references the exiled card. The whole flow is optional at
// resolution: the controller chooses the target when the spell or ability goes on
// the stack, then decides on resolution whether to exile-and-return.
//
// The body compiles to two effects sharing the blink semantics: a leading
// single-target Exile carrying the resolving optionality and a trailing return
// (immediate PutOnBattlefield, or a delayed CreateDelayedTrigger wrapping it)
// whose object binds to the exile's result. This clears the exile's optionality,
// lowers the now-mandatory blink through the ordered effect-sequence path (which
// produces the two-instruction sequence lowerImmediateBlinkReturn or
// lowerDelayedBlinkReturn builds, with the exile rewritten to remember the exiled
// card under a linked key), then marks the exile instruction Optional and
// publishing and gates the return on the exile having succeeded. Declining the
// exile publishes a not-accepted result, so the gated return is skipped and
// nothing returns; accepting it exiles the target and returns it, exactly
// honoring the controller's choice on both branches.
//
// It fails closed (ok=false) unless the body is exactly this controller
// optional-exile-then-return shape lowering to one non-modal, no-shared-target
// two-instruction sequence: a body-level optional, a modal body, a non-single
// ability target, a non-optional or negated/delayed exile, a non-controller
// exile, an independently-optional return, or any lowering that does not produce
// the exact two-instruction blink sequence all leave the body unsupported rather
// than lowered to a silently-wrong sequence.
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
	// The trailing return is either the immediate ", then return that card to the
	// battlefield" blink form or the delayed "If you do, return that card … at the
	// beginning of the next end step" form (Sentinel of the Pearl Trident, Astral
	// Slide, Astral Drift). Either way it must not carry independent optionality:
	// its optionality rides the same resolving "you may" as the exile. The two
	// shapes differ only in whether the put returns this resolution or at the next
	// end step; markBlinkExileOptional gates whichever instruction the ordered
	// sequence emits on the exile result.
	immediateReturn := ret.Connection == parser.EffectConnectionThen && ret.DelayedTiming == 0
	delayedReturn := ret.DelayedTiming == game.DelayedAtBeginningOfNextEndStep
	if ret.Kind != compiler.EffectReturn ||
		ret.Optional ||
		ret.Negated ||
		ret.ToZone != zone.Battlefield ||
		!immediateReturn && !delayedReturn {
		return game.AbilityContent{}, false
	}
	// Clear the exile's resolving optionality and lower the now-mandatory blink
	// through the ordered effect-sequence path, which links the exile to the
	// return and validates that every target and reference is consumed.
	stripped := ctx
	stripped.content.Effects = slices.Clone(ctx.content.Effects)
	stripped.content.Effects[0].Optional = false
	stripped.content.Effects[0].OptionalSpan = shared.Span{}
	// The delayed "If you do, return …" form gates the return on the exile via a
	// resolving-success condition. Drop it so the now-mandatory blink lowers as a
	// plain exile-then-delayed-return; markBlinkExileOptional re-applies the gate
	// on the exile instruction, faithfully reconstructing the "if you do" flow.
	if delayedReturn {
		stripped.content.Conditions = nil
	}
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
	// The return instruction is either the immediate PutOnBattlefield blink or the
	// delayed CreateDelayedTrigger that wraps it; both gate identically on the
	// exile result so declining the exile skips the return entirely.
	if put.Primitive == nil ||
		(put.Primitive.Kind() != game.PrimitivePutOnBattlefield &&
			put.Primitive.Kind() != game.PrimitiveCreateDelayedTrigger) ||
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
