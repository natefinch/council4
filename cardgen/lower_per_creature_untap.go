package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

const perCreatureUntapCountKey = game.ResultKey("per-creature-untap-count")

// lowerEventPlayerPerCreatureUntapPayment lowers the "At the beginning of each
// player's upkeep, that player may choose any number of tapped <filter>
// creatures they control and pay {N} for each creature chosen this way. If the
// player does, untap those creatures." resolution (Dream Tides, Magnetic
// Mountain, Thelon's Curse). The parser folds the per-creature offer onto the
// untap effect as an EffectPaymentFormPerChosenCreature payment carrying the
// event-player payer, the fixed per-creature mana cost, and the creature filter;
// the affirmative "If the player does" gate is a
// ConditionPredicatePriorInstructionAccepted clause.
//
// It lowers to a PayRepeatedly that offers the upkeep player the fixed cost any
// number of times and publishes the count, followed by an Untap that lets the
// same player choose up to that many of their own tapped filtered creatures to
// untap. The untap is gated on the payment having succeeded so a zero-payment
// upkeep resolves to nothing. It fails closed on any shape the runtime cannot
// represent, leaving the effect unsupported rather than lowering a wrong
// sequence.
func lowerEventPlayerPerCreatureUntapPayment(
	_ string,
	ctx contentCtx,
	_ *parser.Ability,
) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Conditions) != 1 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Targets) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	payment := effect.Payment
	condition := ctx.content.Conditions[0]
	if effect.Kind != compiler.EffectUntap ||
		effect.Optional ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		len(effect.Targets) != 0 ||
		payment.Form != parser.EffectPaymentFormPerChosenCreature ||
		payment.Payer != parser.EffectPaymentPayerEventPlayer ||
		len(payment.ManaCost) == 0 ||
		payment.AdditionalCost != nil ||
		payment.GenericManaAmount.DynamicKind != compiler.DynamicAmountNone ||
		manaCostHasVariableSymbol(payment.ManaCost) ||
		condition.Kind != compiler.ConditionIf ||
		condition.Negated ||
		condition.Predicate != compiler.ConditionPredicatePriorInstructionAccepted ||
		condition.NodeID != payment.SuccessConditionNodeID {
		return game.AbilityContent{}, false
	}

	selection, ok := SelectionForSelector(payment.PerCreatureSelector)
	if !ok || selection.Tapped != game.TriTrue {
		return game.AbilityContent{}, false
	}

	resolutionPayment, ok := lowerEventPlayerResolutionPayment(payment)
	if !ok {
		return game.AbilityContent{}, false
	}

	sequence := []game.Instruction{
		{
			Primitive: game.PayRepeatedly{
				Payment:      resolutionPayment,
				PublishCount: perCreatureUntapCountKey,
				Prompt:       resolutionPayment.Prompt,
			},
			PublishResult: perCreatureUntapCountKey,
		},
		{
			Primitive: game.Untap{
				Group: game.PlayerControlledGroup(game.EventPlayerReference(), selection),
				Amount: game.Dynamic(game.DynamicAmount{
					Kind:      game.DynamicAmountChosenNumber,
					ResultKey: perCreatureUntapCountKey,
				}),
				ChooseUpTo: true,
				Chooser:    game.EventPlayerReference(),
			},
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:       perCreatureUntapCountKey,
				Succeeded: game.TriTrue,
			}),
		},
	}
	return game.Mode{Sequence: sequence}.Ability(), true
}
