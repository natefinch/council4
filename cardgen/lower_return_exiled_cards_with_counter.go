package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// lowerReturnExiledCardsWithCounterContent lowers the mass return "Put all
// exiled cards you own with <kind> counters on them into your hand." (Flamewar,
// Brash Veteran) into a single ReturnExiledCardsWithCounter primitive scoped to
// the controller and filtered by the named marker counter the compiler carried
// on the effect. It is the return companion to the exile-with-named-counter
// substrate: the counter kind is read from the typed effect, never from card
// text, so any card that exiles cards under a named marker counter (croak,
// intel, void, collection, ...) lowers through this one path.
//
// It returns ok=false for any shape it does not fully consume: a target,
// reference, condition, mode, or keyword rider, an optional or negated effect, a
// non-controller context, an unknown counter, or a source/destination other
// than exile-to-hand, so an unmodeled wording fails closed.
func lowerReturnExiledCardsWithCounterContent(ctx contentCtx) (game.AbilityContent, bool) {
	// lowerContent calls this only from its len(Effects)==1 block, so a different
	// effect count is a dispatch bug rather than an unsupported card.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerReturnExiledCardsWithCounterContent: reached with %d effects; lowerContent dispatches here only for single-effect content", len(ctx.content.Effects)))
	}
	if ctx.optional ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	// The clause's own "them" pronoun is registered as an internal reference to
	// the exiled cards the effect returns; it is consumed here, not bound to an
	// external antecedent. Any other reference is an unmodeled rider, so fail
	// closed on it.
	for i := range ctx.content.References {
		if !returnExiledCardsWithCounterInternalReference(ctx.content.Effects, ctx.content.References[i]) {
			return game.AbilityContent{}, false
		}
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectReturnExiledCardsWithCounter ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Context != parser.EffectContextController ||
		effect.FromZone != zone.Exile ||
		effect.ToZone != zone.Hand ||
		!effect.CounterKindKnown {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ReturnExiledCardsWithCounter{
			Player:  game.ControllerReference(),
			Counter: effect.CounterKind,
		},
	}}}.Ability(), true
}
