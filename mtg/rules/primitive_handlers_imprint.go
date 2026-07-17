package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// handleReplaceLinkedExiledCard establishes one current object-scoped imprint.
// The prior imprint is disposed only after the new card actually reaches exile,
// so a replacement or stale event reference cannot erase it.
func handleReplaceLinkedExiledCard(r *effectResolver, prim game.ReplaceLinkedExiledCard) effectResolved {
	res := effectResolved{accepted: true}
	cardID, fromZone, ok := resolveCardReference(r.game, r.obj, prim.Card)
	if !ok || fromZone != prim.FromZone {
		return res
	}
	card, ok := r.game.GetCardInstance(cardID)
	if !ok {
		return res
	}
	if !moveCardBetweenZonesWithPlacement(r.game, card.Owner, cardID, fromZone, zone.Exile, false) ||
		!r.game.Players[card.Owner].Exile.Contains(cardID) {
		return res
	}

	key := linkedObjectByObjectKey(r.game, r.obj, string(prim.LinkID))
	for _, ref := range linkedObjects(r.game, key) {
		old, ok := linkedExiledCard(r.game, ref)
		if !ok || old.ID == cardID {
			continue
		}
		moveCardBetweenZonesWithPlacement(r.game, old.Owner, old.ID, zone.Exile, zone.Graveyard, false)
	}
	clearLinkedObjects(r.game, key)
	rememberLinkedObject(r.game, key, game.LinkedObjectRef{
		CardID:          cardID,
		CardZoneVersion: card.ZoneVersion,
	})
	res.succeeded = true
	return res
}
