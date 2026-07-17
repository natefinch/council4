package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerDigRouteSequence lowers the closed look-and-route dig family "Look at the
// top N cards of your library. Put one of them into your hand, put one of them
// on the bottom of your library, and exile one of them. You may play the exiled
// card this turn." (Expressive Iteration) into a single Dig with ordered slots.
// The parser folds all three sentences into one EffectDig marked DigRouteSequence
// whose DigRoute records the look count and the ordered hand / library-bottom /
// exile routes. This text-blind lowerer reads only those typed fields, mapping
// the primary hand route to the Dig's Take and the remaining routes to ordered
// DigSlots (the exile slot carrying the this-turn impulse play grant). Any shape
// the Dig cannot model exactly fails closed.
func lowerDigRouteSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 1 || ctx.optional ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectDig || !effect.DigRouteSequence ||
		effect.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	dig, ok := buildDigRoute(effect.DigRoute)
	if !ok {
		return game.AbilityContent{}, false
	}
	// The look-and-route clauses' only references are the routing sentence's "of
	// them" / "the exiled card" anaphors back to the looked-at cards, which the
	// Dig slots model directly. Every content reference must fall within the
	// consolidated effect's span so no reference needing its own instruction is
	// dropped.
	for ri := range ctx.content.References {
		if !spanCovered(ctx.content.References[ri].Span, []shared.Span{effect.Span}) {
			return game.AbilityContent{}, false
		}
	}
	return game.Mode{
		Sequence: []game.Instruction{{Primitive: dig}},
	}.Ability(), true
}

// buildDigRoute turns a recognized look-and-route payload into a Dig with ordered
// slots. It admits exactly the modeled three-way shape: an ordered hand route
// (the primary Take), a library-bottom route, and an exile route granting play
// this turn, whose counts partition the looked-at cards. Any other destination,
// ordering, count, or grant fails closed.
func buildDigRoute(route parser.DigRouteSyntax) (game.Dig, bool) {
	if route.Look < 1 || len(route.Slots) != 3 {
		return game.Dig{}, false
	}
	hand := route.Slots[0]
	bottom := route.Slots[1]
	exile := route.Slots[2]
	if hand.Destination != zone.Hand || hand.Bottom || hand.PlayThisTurn || hand.Count < 1 ||
		bottom.Destination != zone.Library || !bottom.Bottom || bottom.PlayThisTurn || bottom.Count < 1 ||
		exile.Destination != zone.Exile || exile.Bottom || !exile.PlayThisTurn || exile.Count != 1 {
		return game.Dig{}, false
	}
	if hand.Count+bottom.Count+exile.Count != route.Look {
		return game.Dig{}, false
	}
	return game.Dig{
		Player:    game.ControllerReference(),
		Look:      game.Fixed(route.Look),
		Take:      game.Fixed(hand.Count),
		Remainder: game.DigRemainderLibraryBottom,
		Slots: []game.DigSlot{
			{Count: game.Fixed(bottom.Count), Destination: zone.Library, Bottom: true},
			{Count: game.Fixed(exile.Count), Destination: zone.Exile, Play: opt.Val(game.ImpulsePlayGrant{Duration: game.DurationThisTurn})},
		},
	}, true
}
