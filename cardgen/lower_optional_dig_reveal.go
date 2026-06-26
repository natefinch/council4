package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerOptionalDigReveal lowers the typed optional-reveal dig:
//
//	Look at the top N cards of your library. You may reveal a [filter] card from
//	among them and put it into your hand. Put the rest on the bottom of your
//	library in any order.
//
// The compiler models the body as four effects: a mandatory EffectDig "look"
// carrying the look count, an optional EffectReveal whose selector carries the
// card filter and whose amount distinguishes the single ("a [filter] card"), the
// bounded ("up to N [filter] cards"), and the any-number ("any number of
// [filter] cards") forms, a mandatory EffectPut of the revealed card(s) into the
// controller's hand, and a mandatory EffectPut of the remainder onto the library
// bottom (or into the graveyard). The internal "from among them" / "it" anaphors
// back to the looked-at cards are the only references and the Dig primitive
// models them directly. This lowers the whole body to one game.Dig whose Filter
// restricts which looked-at cards may be taken, whose TakeUpTo carries the "you
// may" (the controller may take none), and whose Reveal carries the reveal.
//
// It fails closed unless the body is exactly this four-effect shape with a
// representable card filter and a recognized remainder destination: a body-level
// optional, a modal or targeted body, a keyword-bearing body, a non-controller
// subject, a negated effect, a filter cardSelectionForSelector cannot project, an
// independently optional put, or a reference needing its own instruction all
// leave the body unsupported rather than lowering a silently-wrong sequence.
func lowerOptionalDigReveal(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Effects) != 4 {
		return game.AbilityContent{}, false
	}
	look := ctx.content.Effects[0]
	reveal := ctx.content.Effects[1]
	putHand := ctx.content.Effects[2]
	putRest := ctx.content.Effects[3]

	if look.Kind != compiler.EffectDig || !look.Exact || look.Optional || look.Negated ||
		look.Context != parser.EffectContextController ||
		!look.Amount.Known || len(look.Targets) != 0 || len(look.References) != 0 {
		return game.AbilityContent{}, false
	}
	lookCount := look.Amount.Value
	if lookCount < 1 {
		return game.AbilityContent{}, false
	}

	if reveal.Kind != compiler.EffectReveal || !reveal.Optional || reveal.Negated ||
		reveal.Context != parser.EffectContextController || len(reveal.Targets) != 0 {
		return game.AbilityContent{}, false
	}
	filter, ok := cardSelectionForSelector(reveal.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	takeCount, ok := optionalDigTakeCount(reveal.Amount, lookCount)
	if !ok {
		return game.AbilityContent{}, false
	}

	if putHand.Kind != compiler.EffectPut || putHand.Optional || putHand.Negated ||
		putHand.Context != parser.EffectContextController || len(putHand.Targets) != 0 ||
		putHand.ToZone != zone.Hand ||
		putHand.Destination != parser.EffectDestinationUnspecified {
		return game.AbilityContent{}, false
	}

	if putRest.Kind != compiler.EffectPut || putRest.Optional || putRest.Negated ||
		putRest.Context != parser.EffectContextController || len(putRest.Targets) != 0 {
		return game.AbilityContent{}, false
	}
	remainder, ok := optionalDigRemainder(putRest)
	if !ok {
		return game.AbilityContent{}, false
	}

	// The only references in the body are the internal "from among them" / "it"
	// anaphors back to the looked-at cards, which the Dig primitive models
	// directly. Every content reference must fall within one of the four effect
	// spans so no reference needing its own instruction is dropped.
	covering := []shared.Span{look.Span, reveal.Span, putHand.Span, putRest.Span}
	for ri := range ctx.content.References {
		if !spanCovered(ctx.content.References[ri].Span, covering) {
			return game.AbilityContent{}, false
		}
	}

	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.Dig{
			Player:    game.ControllerReference(),
			Look:      game.Fixed(lookCount),
			Take:      game.Fixed(takeCount),
			Remainder: remainder,
			Filter:    opt.Val(filter),
			TakeUpTo:  true,
			Reveal:    true,
		},
	}}}.Ability(), true
}

// optionalDigTakeCount maps the reveal effect's amount to the upper bound on the
// cards the controller may take into their hand. The single form ("a [filter]
// card") carries a known amount of one; the bounded form ("up to N [filter]
// cards") carries a known amount of N; the any-number form ("any number of
// [filter] cards") carries an unknown amount and is bounded by the look count
// (the controller may take every matching looked-at card). Each form is an upper
// bound on an optional reveal, so the controller may always take fewer (down to
// none). A variable-X amount fails closed.
func optionalDigTakeCount(amount compiler.CompiledAmount, lookCount int) (int, bool) {
	switch {
	case amount.VariableX:
		return 0, false
	case amount.Known && amount.Value >= 1:
		return amount.Value, true
	case !amount.Known && !amount.RangeKnown && amount.DynamicKind == compiler.DynamicAmountNone:
		return lookCount, true
	default:
		return 0, false
	}
}

// optionalDigRemainder maps the remainder put clause to the runtime Dig
// remainder: "Put the rest on the bottom of your library ..." routes to the
// library bottom, "... into your graveyard." routes to the graveyard. Any other
// destination fails closed.
func optionalDigRemainder(putRest compiler.CompiledEffect) (game.DigRemainder, bool) {
	switch {
	case putRest.ToZone == zone.Library && putRest.Destination == parser.EffectDestinationBottom:
		return game.DigRemainderLibraryBottom, true
	case putRest.ToZone == zone.Graveyard && putRest.Destination == parser.EffectDestinationUnspecified:
		return game.DigRemainderGraveyard, true
	default:
		return 0, false
	}
}
