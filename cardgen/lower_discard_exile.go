package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// lowerDiscardExileFromGraveyard lowers Necropotence's mandatory discard trigger
// "Whenever you discard a card, exile that card from your graveyard." — a single
// non-optional exile of the just-discarded card out of the graveyard, granting no
// permission (unlike Containment Construct's optional exile-then-play shape, which
// lowerExileForPlay handles).
//
// The discarded card rests in its owner's graveyard when the trigger resolves;
// the runtime resolves the "that card" (ReferenceThatObject) back-reference
// through CardReferenceEvent to the exact card the triggering discard event
// carried, so one trigger fires per discarded card — including simultaneous and
// multi-card discards, each of which emits its own EventCardDiscarded — and each
// resolution exiles only its own card. The move no-ops if that card has already
// left the graveyard (a zone-version mismatch, e.g. another effect moved it
// first), matching "exile that card from your graveyard" doing nothing when the
// card is no longer there. It lowers to a bare MoveCard from graveyard to exile,
// which relocates the card under its owner, so a control or source change to the
// trigger never misroutes the exile.
func lowerDiscardExileFromGraveyard(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.enclosingKind != compiler.AbilityTriggered ||
		ctx.triggerEvent != game.EventCardDiscarded ||
		ctx.triggerOneOrMore ||
		ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Conditions) != 0 {
		return game.AbilityContent{}, false
	}
	exile := ctx.content.Effects[0]
	if exile.Kind != compiler.EffectExile ||
		exile.Optional ||
		exile.Negated ||
		exile.Context != parser.EffectContextController ||
		exile.FromZone != zone.Graveyard ||
		exile.DelayedTiming != 0 ||
		exile.CounterKindKnown ||
		len(exile.Targets) != 0 ||
		!exileBackReferencesSingleObject(exile) {
		return game.AbilityContent{}, false
	}
	consumed := ctx
	consumed.content.References = nil
	if consumed.content.Unconsumed() {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.MoveCard{
			Card:        game.CardReference{Kind: game.CardReferenceEvent},
			FromZone:    zone.Graveyard,
			Destination: zone.Exile,
		},
	}}}.Ability(), true
}
