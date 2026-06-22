package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// milledCardsLinkKey labels the milled cards a "mill, then optionally pay to
// keep one" sequence publishes so the gated return can restrict itself to
// exactly those cards.
const milledCardsLinkKey = game.LinkedKey("milled-cards")

// lowerMillThenPaidReturnSequence lowers the ordered sequence "mill N cards.
// Then you may pay <cost>. If you do, put a card from among those cards into
// your hand." (Ripples of Undeath). The mill is mandatory and publishes the
// cards it milled; the optional combined payment gates a follow-up that returns
// one of exactly those milled cards from the graveyard to the controller's
// hand. It keys entirely on the typed effect shape — a mandatory controller
// mill followed by an exact controller "those cards" put carrying a resolution
// payment and a prior-instruction-accepted gate — so it stays text-blind and
// fails closed on any other sequence.
func lowerMillThenPaidReturnSequence(cardName string, ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 2 ||
		len(ctx.content.Conditions) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 1 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	mill := ctx.content.Effects[0]
	put := ctx.content.Effects[1]
	if mill.Kind != compiler.EffectMill ||
		mill.Context != parser.EffectContextController ||
		!mill.Exact ||
		mill.Negated ||
		mill.Optional ||
		mill.DelayedTiming != 0 ||
		!mill.Amount.Known ||
		mill.Amount.Value < 1 ||
		len(mill.References) != 0 {
		return game.AbilityContent{}, false
	}
	if put.Kind != compiler.EffectPut ||
		put.Context != parser.EffectContextController ||
		!put.Exact ||
		put.Negated ||
		put.Optional ||
		put.DelayedTiming != 0 ||
		!put.Amount.Known ||
		put.Amount.Value != 1 ||
		len(put.References) != 1 ||
		put.References[0].Pronoun != compiler.ReferencePronounThose {
		return game.AbilityContent{}, false
	}
	payment := put.Payment
	condition := ctx.content.Conditions[0]
	if payment.Form != parser.EffectPaymentFormMayPayThenIfDo ||
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
	millAmount, ok := cardCountQuantity(mill.Amount, false)
	if !ok {
		return game.AbilityContent{}, false
	}
	sequence := []game.Instruction{
		{Primitive: game.Mill{
			Amount:        millAmount,
			Player:        game.ControllerReference(),
			PublishLinked: milledCardsLinkKey,
		}},
		{
			Primitive:     game.Pay{Payment: resolutionPayment},
			PublishResult: controllerPaidResultKey,
		},
		{
			Primitive: game.ReturnFromGraveyard{
				Player:      game.ControllerReference(),
				Amount:      game.Fixed(1),
				Destination: zone.Hand,
				FromLinked:  milledCardsLinkKey,
			},
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:       controllerPaidResultKey,
				Succeeded: game.TriTrue,
			}),
		},
	}
	return game.Mode{Sequence: sequence}.Ability(), true
}
