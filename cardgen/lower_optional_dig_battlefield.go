package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerOptionalDigToBattlefield lowers the typed optional dig-to-battlefield:
//
//	Look at the top N cards of your library. You may put a [filter] card from
//	among them onto the battlefield [tapped]. Put the rest <remainder>.
//
// The compiler models the body as three effects: a mandatory EffectDig "look"
// carrying the look count, an optional EffectPut whose destination is the
// battlefield and whose selector carries the card filter (its amount
// distinguishes the single "a [filter] card", the bounded "up to N [filter]
// cards", and the any-number "any number of [filter] cards" forms), and a
// mandatory EffectPut of the remainder onto the library bottom or into the
// graveyard. The internal "from among them" / "the rest" anaphors back to the
// looked-at cards are the only references and the Dig primitive models them
// directly. This lowers the whole body to one game.Dig whose Filter restricts
// which looked-at cards may be put onto the battlefield, whose TakeUpTo carries
// the "you may" (the controller may put none), and whose Destination is the
// battlefield (EntersTapped carrying the "tapped" entry).
//
// It fails closed unless the body is exactly this three-effect shape with a
// representable card filter and a recognized remainder destination: a body-level
// optional, a modal/targeted/keyword-bearing/conditional body, a non-controller
// subject, a negated effect, a variable look count, an attacking/blocking entry
// rider the flat filter cannot carry, a filter cardSelectionForSelector cannot
// project, an independently optional remainder, or a reference needing its own
// instruction all leave the body unsupported rather than lowering a
// silently-wrong sequence.
func lowerOptionalDigToBattlefield(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Effects) != 3 {
		return game.AbilityContent{}, false
	}
	look := ctx.content.Effects[0]
	put := ctx.content.Effects[1]
	putRest := ctx.content.Effects[2]

	if look.Kind != compiler.EffectDig || !look.Exact || look.Optional || look.Negated ||
		look.Context != parser.EffectContextController ||
		!look.Amount.Known || len(look.Targets) != 0 || len(look.References) != 0 {
		return game.AbilityContent{}, false
	}
	lookCount := look.Amount.Value
	if lookCount < 1 {
		return game.AbilityContent{}, false
	}

	if put.Kind != compiler.EffectPut || !put.Optional || put.Negated ||
		put.Context != parser.EffectContextController || len(put.Targets) != 0 ||
		put.ToZone != zone.Battlefield ||
		put.Destination != parser.EffectDestinationUnspecified {
		return game.AbilityContent{}, false
	}
	// A "tapped and attacking" / "tapped and blocking" entry carries combat-state
	// riders the flat Dig has no field for, so the parser leaves them only on the
	// selector. Reject them rather than silently dropping the attacking/blocking
	// entry; the plain "tapped" entry is modeled by EntersTapped below.
	if put.Selector.Attacking || put.Selector.Blocking {
		return game.AbilityContent{}, false
	}
	filter, ok := cardSelectionForSelector(put.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	takeCount, ok := optionalDigTakeCount(put.Amount, lookCount)
	if !ok {
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

	// The only references in the body are the internal "from among them" / "the
	// rest" anaphors back to the looked-at cards, which the Dig primitive models
	// directly. Every content reference must fall within one of the three effect
	// spans so no reference needing its own instruction is dropped.
	covering := []shared.Span{look.Span, put.Span, putRest.Span}
	for ri := range ctx.content.References {
		if !spanCovered(ctx.content.References[ri].Span, covering) {
			return game.AbilityContent{}, false
		}
	}

	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.Dig{
			Player:       game.ControllerReference(),
			Look:         game.Fixed(lookCount),
			Take:         game.Fixed(takeCount),
			Remainder:    remainder,
			Filter:       opt.Val(filter),
			TakeUpTo:     true,
			Destination:  zone.Battlefield,
			EntersTapped: put.EntersTapped,
		},
	}}}.Ability(), true
}
