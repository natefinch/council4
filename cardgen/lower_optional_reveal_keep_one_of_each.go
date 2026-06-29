package cardgen

import (
	"reflect"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerOptionalRevealKeepOneOfEach lowers the typed optional one-of-each reveal
// keep:
//
//	Reveal the top N cards of your library. You may put a [type-A] card and/or a
//	[type-B] card from among them into your hand. Put the rest into your
//	graveyard.
//
// The compiler models the body as three effects: a mandatory EffectReveal of the
// top N cards, an optional EffectPut into hand whose selector is the inclusive
// one-of-each union of named card types (each at most one), and a mandatory
// EffectPut of the remainder into the graveyard. Unlike the single-filter
// lowerOptionalRevealTakeToGraveyard, the take put names two or more types joined
// by "and/or", so the controller may take up to one card of each type rather than
// one card matching either.
//
// The "reveal then put the rest into your graveyard" form is behaviorally a mill
// of all N cards followed by an optional per-type retrieval: the reveal publicly
// shows every card, the kept cards reach hand, and the rest rest in the graveyard
// exactly as a mill leaves them. It lowers to a Mill of N that publishes its
// milled cards, then one independent optional ReturnFromGraveyardChoice per named
// type that returns up to one matching milled card to hand. This mirrors the
// graveyard-equivalent mill keep (lowerOptionalMillKeepHand) and the battlefield
// one-of-each form (lowerMillThenOptionalAmongOneOfEachToBattlefield).
//
// It fails closed unless the body is exactly this three-effect shape with two or
// more named pure card-type filters and a graveyard remainder: a body-level
// optional, a modal, targeted, or keyword-bearing body, a non-controller subject,
// a negated or delayed effect, a non-singular per-type take, a non-graveyard
// remainder, a filter that is not the inclusive one-of-each union, or a reference
// needing its own instruction all leave the body unsupported.
func lowerOptionalRevealKeepOneOfEach(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Effects) != 3 {
		return game.AbilityContent{}, false
	}
	reveal := ctx.content.Effects[0]
	putHand := ctx.content.Effects[1]
	putRest := ctx.content.Effects[2]

	if reveal.Kind != compiler.EffectReveal || reveal.Optional || reveal.Negated ||
		reveal.DelayedTiming != 0 ||
		reveal.Context != parser.EffectContextController ||
		!reveal.Amount.Known || reveal.Amount.Value < 1 ||
		len(reveal.Targets) != 0 || len(reveal.References) != 0 {
		return game.AbilityContent{}, false
	}
	revealCount := reveal.Amount.Value

	if putHand.Kind != compiler.EffectPut || !putHand.Optional || putHand.Negated ||
		putHand.DelayedTiming != 0 ||
		putHand.Context != parser.EffectContextController || len(putHand.Targets) != 0 ||
		putHand.ToZone != zone.Hand ||
		putHand.Destination != parser.EffectDestinationUnspecified ||
		!putHand.Selector.InclusiveOneOfEach ||
		!putHand.Amount.Known || putHand.Amount.Value != 1 ||
		putHand.Payment.Form != parser.EffectPaymentFormUnknown {
		return game.AbilityContent{}, false
	}
	selections, ok := oneOfEachHandSelections(putHand.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}

	if putRest.Kind != compiler.EffectPut || putRest.Optional || putRest.Negated ||
		putRest.DelayedTiming != 0 ||
		putRest.Context != parser.EffectContextController || len(putRest.Targets) != 0 {
		return game.AbilityContent{}, false
	}
	remainder, ok := optionalDigRemainder(putRest)
	if !ok || remainder != game.DigRemainderGraveyard {
		return game.AbilityContent{}, false
	}

	covering := []shared.Span{reveal.Span, putHand.Span, putRest.Span}
	for ri := range ctx.content.References {
		if !spanCovered(ctx.content.References[ri].Span, covering) {
			return game.AbilityContent{}, false
		}
	}

	sequence := []game.Instruction{{Primitive: game.Mill{
		Amount:        game.Fixed(revealCount),
		Player:        game.ControllerReference(),
		PublishLinked: milledCardsLinkKey,
	}}}
	for _, selection := range selections {
		sequence = append(sequence, game.Instruction{
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
		})
	}
	return game.Mode{Sequence: sequence}.Ability(), true
}

// oneOfEachHandSelections splits an inclusive one-of-each card selector ("a
// creature card and/or an enchantment card") into one game.Selection per named
// card type so each can drive an independent optional keep. Unlike
// oneOfEachCardSelections (which requires the bare card selector kind) this keeps
// the union types whichever selector kind anchors them, but accepts only a pure
// union of card types: any color, supertype, mana value, controller, keyword,
// counter, name, subtype, or zone qualifier fails closed so a richer filter never
// silently splits into wrong picks. At least two named types are required for the
// one-of-each wording to be meaningful.
func oneOfEachHandSelections(selector compiler.CompiledSelector) ([]game.Selection, bool) {
	selection, ok := cardSelectionForSelector(selector)
	if !ok {
		return nil, false
	}
	pure := game.Selection{RequiredTypesAny: selection.RequiredTypesAny}
	if !reflect.DeepEqual(selection, pure) || len(selection.RequiredTypesAny) < 2 {
		return nil, false
	}
	selections := make([]game.Selection, 0, len(selection.RequiredTypesAny))
	for _, cardType := range selection.RequiredTypesAny {
		selections = append(selections, game.Selection{RequiredTypes: []types.Card{cardType}})
	}
	return selections, true
}
