package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerOptionalRevealTakeToGraveyard lowers the typed optional-take reveal dig:
//
//	Reveal the top N cards of your library. You may put a [filter] card from
//	among them into your hand. Put the rest into your graveyard.
//
// The compiler models the body as three effects: a mandatory EffectReveal that
// reveals the top N cards (carrying the look count), an optional EffectPut of one
// or more matching cards into the controller's hand (carrying the card filter and
// the take count), and a mandatory EffectPut of the remainder into the
// controller's graveyard. Unlike the look-at form (lowerOptionalDigReveal), the
// optionality and the card filter both ride on the take-into-hand put rather than
// on a separate reveal effect.
//
// It lowers the whole body to one game.Dig whose Filter restricts which revealed
// cards may be taken, whose TakeUpTo carries the "you may" (the controller may
// take none), whose Remainder routes the non-taken cards to the graveyard, and
// whose Reveal carries the reveal. The remainder is required to be the graveyard:
// the Oracle "Reveal the top N cards" publicly reveals every looked-at card, and
// routing the non-taken cards to the graveyard leaves them public exactly as the
// reveal does (and enters them into the graveyard in the same count), so the Dig
// model — which reveals the taken cards and graveyards the rest — is behaviorally
// equivalent. A library-bottom remainder is not equivalent (the bottomed cards
// would have been publicly revealed first) and fails closed.
//
// It fails closed unless the body is exactly this three-effect shape with a
// representable card filter: a body-level optional, a modal, targeted, or
// keyword-bearing body, a non-controller subject, a negated or delayed effect, an
// inclusive one-of-each ("creature card and/or an enchantment card") filter whose
// per-type take count Dig cannot model, a filter cardSelectionForSelector cannot
// project, an independently optional reveal or remainder put, a non-graveyard
// remainder, or a reference needing its own instruction all leave the body
// unsupported rather than lowering a silently-wrong sequence.
func lowerOptionalRevealTakeToGraveyard(ctx contentCtx) (game.AbilityContent, bool) {
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
	lookCount := reveal.Amount.Value

	if putHand.Kind != compiler.EffectPut || !putHand.Optional || putHand.Negated ||
		putHand.DelayedTiming != 0 ||
		putHand.Context != parser.EffectContextController || len(putHand.Targets) != 0 ||
		putHand.ToZone != zone.Hand ||
		putHand.Destination != parser.EffectDestinationUnspecified ||
		putHand.Selector.InclusiveOneOfEach ||
		putHand.Payment.Form != parser.EffectPaymentFormUnknown {
		return game.AbilityContent{}, false
	}
	filter, ok := cardSelectionForSelector(putHand.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	takeCount, ok := optionalDigTakeCount(putHand.Amount, lookCount)
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

	// The only references in the body are the internal "from among them" / "the
	// rest" anaphors back to the revealed cards, which the Dig primitive models
	// directly. Every content reference must fall within one of the three effect
	// spans so no reference needing its own instruction is dropped.
	covering := []shared.Span{reveal.Span, putHand.Span, putRest.Span}
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
