package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/opt"
)

const variablePayScaledDrawCountKey = game.ResultKey("variable-pay-scaled-draw-count")

// lowerControllerVariablePayScaledDraw lowers the "Whenever you gain life, you
// may pay {X}, where X is less than or equal to the amount of life you gained. If
// you do, draw X cards." trigger body (Well of Lost Dreams). The parser folds the
// bounded optional payment onto the draw effect as an
// EffectPaymentFormMayPayVariableUpTo payment carrying the controller payer and
// the triggering life-change bound in GenericManaAmount; the affirmative "If you
// do" gate is a ConditionPredicatePriorInstructionAccepted clause, and the draw
// amount is the chosen variable X.
//
// It lowers to a PayRepeatedly that offers the controller {1} up to the amount of
// life gained times and publishes the count, followed by a Draw of that many
// cards. The Draw is gated on the payment having succeeded so a zero payment
// resolves to no draw. Modeling the single "{X}" payment as {1} paid X times is
// behaviorally identical under generic mana — each unit paid draws one card, and
// the published count is the chosen X — while reusing the count-publish/scaled
// effect pipeline. Any shape the runtime sequence cannot represent fails closed,
// leaving the body unsupported rather than lowering a wrong sequence.
func lowerControllerVariablePayScaledDraw(
	_ string,
	ctx contentCtx,
	_ *parser.Ability,
) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Conditions) != 1 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}
	if ctx.triggerEvent != game.EventLifeGained {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	payment := effect.Payment
	condition := ctx.content.Conditions[0]
	if effect.Kind != compiler.EffectDraw ||
		effect.Optional ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		len(effect.Targets) != 0 ||
		effect.Context != parser.EffectContextController ||
		!effect.Amount.VariableX ||
		effect.Amount.Known ||
		effect.Amount.DynamicKind != compiler.DynamicAmountNone ||
		payment.Form != parser.EffectPaymentFormMayPayVariableUpTo ||
		payment.Payer != parser.EffectPaymentPayerController ||
		len(payment.ManaCost) != 0 ||
		payment.AdditionalCost != nil ||
		payment.GenericManaAmount.DynamicKind != compiler.DynamicAmountTriggeringLifeChange ||
		condition.Kind != compiler.ConditionIf ||
		condition.Negated ||
		condition.Predicate != compiler.ConditionPredicatePriorInstructionAccepted ||
		condition.NodeID != payment.SuccessConditionNodeID {
		return game.AbilityContent{}, false
	}

	bound, ok := lowerTriggeringEventQuantityAmount(ctx, payment.GenericManaAmount)
	if !ok || bound.Kind != game.DynamicAmountEventLifeChange {
		return game.AbilityContent{}, false
	}
	paidCount := game.Dynamic(game.DynamicAmount{
		Kind:      game.DynamicAmountChosenNumber,
		ResultKey: variablePayScaledDrawCountKey,
	})

	sequence := []game.Instruction{
		{
			Primitive: game.PayRepeatedly{
				Payment: game.ResolutionPayment{
					Prompt:   "Pay {1} to draw a card?",
					Payer:    opt.Val(game.ControllerReference()),
					ManaCost: opt.Val(cost.Mana{cost.O(1)}),
				},
				PublishCount: variablePayScaledDrawCountKey,
				MaxCount:     opt.Val(&bound),
			},
			PublishResult: variablePayScaledDrawCountKey,
		},
		{
			Primitive: game.Draw{
				Amount: paidCount,
				Player: game.ControllerReference(),
			},
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:       variablePayScaledDrawCountKey,
				Succeeded: game.TriTrue,
			}),
		},
	}
	return game.Mode{Sequence: sequence}.Ability(), true
}
