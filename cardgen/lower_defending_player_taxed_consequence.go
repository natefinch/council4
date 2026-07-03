package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

const defendingPlayerUnpaidResultKey = game.ResultKey("defending-player-unpaid")

// lowerDefendingPlayerTaxedSourceConsequence lowers the attack-triggered
// defending-player optional-payment failure gate "defending player may pay {N}.
// If that player doesn't, <consequence>." (Shrouded Serpent). It is the
// defending-player, source-consequence counterpart of
// lowerEventPlayerTaxedControllerBenefit: the player being attacked is offered
// the payment, and when they decline the consequence resolves.
//
// The consequence is lowered compositionally through the shared content path
// after stripping the payment and the "if that player doesn't" gate, so it fails
// closed on any consequence the backend cannot already lower ("this creature
// can't be blocked this turn." lowers through lowerCantBeBlockedSpell). Each
// resulting instruction is gated on the payment having NOT succeeded
// (TriFalse), and a resolution Pay charged to the defending player is prepended
// to publish that result. Only the consequence body's own references are handed
// to the body lowering; the gate's "that player" payer reference sits ahead of
// the body span and is dropped, so a source-subject consequence sees only its
// own source reference.
func lowerDefendingPlayerTaxedSourceConsequence(
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
		if ctx.content.Effects[i].Payment.Payer == parser.EffectPaymentPayerDefendingPlayer {
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
		payment.Form != parser.EffectPaymentFormMayPayThenIfDoesNot ||
		payment.Payer != parser.EffectPaymentPayerDefendingPlayer ||
		manaCostHasVariableSymbol(payment.ManaCost) ||
		payment.GenericManaAmount.DynamicKind != compiler.DynamicAmountNone ||
		condition.Kind != compiler.ConditionIf ||
		condition.Predicate != compiler.ConditionPredicateDefendingPlayerDoesNotPay ||
		condition.NodeID != payment.FailureConditionNodeID ||
		payment.Span.End.Offset >= condition.Span.Start.Offset {
		return game.AbilityContent{}, false
	}

	// Keep only references that sit in the consequence body, past the "if that
	// player doesn't" gate. The payment payer ("that player") reference lives
	// inside the gate span and must not leak into the source-subject body, which
	// lowers exactly one source reference.
	bodyReferences := make([]compiler.CompiledReference, 0, len(ctx.content.References))
	for _, reference := range ctx.content.References {
		if reference.Span.Start.Offset >= condition.Span.End.Offset {
			bodyReferences = append(bodyReferences, reference)
		}
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
			Key:       defendingPlayerUnpaidResultKey,
			Succeeded: game.TriFalse,
		})
	}
	resolutionPayment := game.ResolutionPayment{
		Prompt:   "Pay " + payment.ManaCost.String() + "?",
		Payer:    opt.Val(game.DefendingPlayerReference()),
		ManaCost: opt.Val(slices.Clone(payment.ManaCost)),
	}
	sequence := make([]game.Instruction, 0, len(consequence)+1)
	sequence = append(sequence, game.Instruction{
		Primitive:     game.Pay{Payment: resolutionPayment},
		PublishResult: defendingPlayerUnpaidResultKey,
	})
	sequence = append(sequence, consequence...)
	return game.Mode{Sequence: sequence}.Ability(), true
}
