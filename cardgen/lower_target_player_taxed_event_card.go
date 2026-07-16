package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

const targetPlayerPaidResultKey = game.ResultKey("target-player-paid")

// lowerTargetPlayerTaxedEventCardEffect composes an ordinary event-card effect
// with an "unless target player pays <cost>" offer. The player target is chosen
// when the ability is put on the stack, pays during resolution, and the event-card
// effect runs only when that payment fails.
func lowerTargetPlayerTaxedEventCardEffect(cardName string, ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Conditions) != 1 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	payment := effect.Payment
	condition := ctx.content.Conditions[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.DelayedTiming != 0 ||
		payment.Form != parser.EffectPaymentFormUnless ||
		payment.Payer != parser.EffectPaymentPayerTargetPlayer ||
		len(payment.ManaCost) != 0 ||
		payment.AdditionalCost == nil ||
		payment.GenericManaAmount.DynamicKind != compiler.DynamicAmountNone ||
		condition.Kind != compiler.ConditionUnless ||
		condition.Predicate != compiler.ConditionPredicateTargetControllerDoesNotPay ||
		!condition.Order.Contains(payment.Order) {
		return game.AbilityContent{}, false
	}
	target, ok := playerTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	resolutionPayment, ok := controllerPaidResolutionPayment(cardName, payment)
	if !ok {
		return game.AbilityContent{}, false
	}
	resolutionPayment.Payer = opt.Val(game.TargetPlayerReference(0))

	stripped := ctx
	stripped.content.Conditions = nil
	stripped.content.Targets = nil
	effect.Payment = compiler.CompiledEffectPayment{}
	effect.Targets = nil
	stripped.content.Effects = []compiler.CompiledEffect{effect}
	benefit, ok := lowerEventCardEffect(stripped)
	if !ok ||
		len(benefit.SharedTargets) != 0 ||
		len(benefit.Modes) != 1 ||
		len(benefit.Modes[0].Targets) != 0 ||
		len(benefit.Modes[0].Sequence) != 1 {
		return game.AbilityContent{}, false
	}
	instruction := benefit.Modes[0].Sequence[0]
	instruction.ResultGate = opt.Val(game.InstructionResultGate{
		Key:       targetPlayerPaidResultKey,
		Succeeded: game.TriFalse,
	})
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{
			{
				Primitive:     game.Pay{Payment: resolutionPayment},
				PublishResult: targetPlayerPaidResultKey,
			},
			instruction,
		},
	}.Ability(), true
}
