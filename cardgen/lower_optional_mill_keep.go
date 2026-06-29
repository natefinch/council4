package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerOptionalMillKeepHand lowers the typed optional-take mill:
//
//	Mill N cards. You may put a [filter] card from among the cards milled this
//	way into your hand.
//
// The compiler models the body as two effects: a mandatory EffectMill that puts
// the top N cards from the controller's library into their graveyard (carrying
// the mill count), and an optional EffectPut of one or more matching milled cards
// into the controller's hand (carrying the card filter and take count). Unlike
// the reveal/look digs there is no explicit "put the rest" clause: the cards the
// controller declines simply stay in the graveyard where the mill placed them.
//
// It lowers the whole body to one game.Dig whose Filter restricts which milled
// cards may be taken, whose TakeUpTo carries the "you may" (the controller may
// take none), and whose graveyard remainder leaves the non-taken cards where the
// mill put them. The mill publicly moves every card into the graveyard, so the
// Dig model — which reveals the taken cards and graveyards the rest — is
// behaviorally equivalent: the same N cards leave the top of the library, the
// same non-taken cards rest in the graveyard, and the taken cards reach hand.
//
// It fails closed unless the body is exactly this two-effect shape with a
// representable card filter: a body-level optional, a modal, targeted, or
// keyword-bearing body, a non-controller subject, a negated or delayed effect, an
// inclusive one-of-each filter Dig cannot model, a take amount tied to payment, a
// filter cardSelectionForSelector cannot project, or a reference needing its own
// instruction all leave the body unsupported rather than lowering a silently-
// wrong sequence.
func lowerOptionalMillKeepHand(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Effects) != 2 {
		return game.AbilityContent{}, false
	}
	mill := ctx.content.Effects[0]
	putHand := ctx.content.Effects[1]

	if mill.Kind != compiler.EffectMill || mill.Optional || mill.Negated ||
		mill.DelayedTiming != 0 ||
		mill.Context != parser.EffectContextController ||
		!mill.Amount.Known || mill.Amount.Value < 1 ||
		len(mill.Targets) != 0 || len(mill.References) != 0 {
		return game.AbilityContent{}, false
	}
	millCount := mill.Amount.Value

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
	takeCount, ok := optionalDigTakeCount(putHand.Amount, millCount)
	if !ok {
		return game.AbilityContent{}, false
	}

	// The only references in the body are the internal "from among the cards
	// milled this way" anaphor back to the milled cards, which the Dig primitive
	// models directly. Every content reference must fall within one of the two
	// effect spans so no reference needing its own instruction is dropped.
	covering := []shared.Span{mill.Span, putHand.Span}
	for ri := range ctx.content.References {
		if !spanCovered(ctx.content.References[ri].Span, covering) {
			return game.AbilityContent{}, false
		}
	}

	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.Dig{
			Player:    game.ControllerReference(),
			Look:      game.Fixed(millCount),
			Take:      game.Fixed(takeCount),
			Remainder: game.DigRemainderGraveyard,
			Filter:    opt.Val(filter),
			TakeUpTo:  true,
			Reveal:    true,
		},
	}}}.Ability(), true
}
