package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerMillThenOptionalAmongToHandSequence lowers the ordered sequence "mill N
// cards. You may put a <filter> card from among them into your hand." (Six's
// attack trigger). The mill is mandatory and publishes the cards it milled; the
// optional "you may put" returns one of exactly those milled cards from the
// graveyard to the controller's hand, restricted to the put clause's printed
// card filter ("a land card"). Unlike lowerMillThenPaidReturnSequence the
// follow-up carries no payment: its optionality is the plain "you may" the
// runtime offers the controller.
//
// It keys entirely on the typed effect shape — a mandatory controller mill
// followed by an optional controller "them" put into hand with no resolution
// payment — so it stays text-blind and fails closed on any other sequence.
func lowerMillThenOptionalAmongToHandSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 2 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Targets) != 0 ||
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
		len(mill.References) != 0 ||
		len(mill.Targets) != 0 {
		return game.AbilityContent{}, false
	}
	if put.Kind != compiler.EffectPut ||
		put.Context != parser.EffectContextController ||
		!put.Optional ||
		put.Negated ||
		put.DelayedTiming != 0 ||
		put.ToZone != zone.Hand ||
		put.Payment.Form != parser.EffectPaymentFormUnknown ||
		!put.Amount.Known ||
		put.Amount.RangeKnown ||
		put.Amount.VariableX ||
		put.Amount.Value != 1 ||
		len(put.Targets) != 0 ||
		len(put.References) != 1 ||
		put.References[0].Pronoun != compiler.ReferencePronounThem {
		return game.AbilityContent{}, false
	}
	selection, ok := cardSelectionForSelector(put.Selector)
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
			Primitive: game.ReturnFromGraveyardChoice(
				game.ControllerReference(),
				selection,
				game.Fixed(1),
				zone.Hand,
				false,
				opt.V[int]{},
				false,
				milledCardsLinkKey,
			),
			Optional: true,
		},
	}
	return game.Mode{Sequence: sequence}.Ability(), true
}
