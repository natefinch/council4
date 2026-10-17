package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// handlePutLinkedExiledCardsInLibrary disposes of every card a sibling
// exile-until-leaves clause exiled under prim.LinkedKey by moving it from exile
// to its owner's library, to the bottom when prim.Bottom is set. It backs the
// linked disposal "The owner of each card exiled with <this permanent> puts that
// card on the bottom of their library." (Trial of a Time Lord's guilty verdict),
// consuming the link so the synthesized leaves-the-battlefield trigger finds
// nothing left to return. Each disposed card is matched by the linked object
// identity, so an object that already left exile is skipped.
func handlePutLinkedExiledCardsInLibrary(r *effectResolver, prim game.PutLinkedExiledCardsInLibrary) effectResolved {
	res := effectResolved{accepted: true}
	key := linkedObjectSourceKey(r.game, r.obj, string(prim.LinkedKey))
	refs := linkedObjects(r.game, key)
	if prim.RandomOrder {
		cardsByOwner := make(map[game.PlayerID][]game.ZoneCardSnapshot)
		var owners []game.PlayerID
		for _, ref := range refs {
			card, ok := linkedExiledCard(r.game, ref)
			if !ok {
				continue
			}
			if _, seen := cardsByOwner[card.Owner]; !seen {
				owners = append(owners, card.Owner)
			}
			cardsByOwner[card.Owner] = append(cardsByOwner[card.Owner], game.ZoneCardSnapshot{
				CardID: card.ID, ZoneVersion: card.ZoneVersion,
			})
		}
		for _, ownerID := range owners {
			owner, ok := playerByID(r.game, ownerID)
			if !ok {
				continue
			}
			if bottomExiledCards(r.game, owner, ownerID, cardsByOwner[ownerID], r.engine.rng) {
				res.succeeded = true
			}
		}
		clearLinkedObjects(r.game, key)
		return res
	}
	for _, ref := range refs {
		card, ok := linkedExiledCard(r.game, ref)
		if !ok {
			continue
		}
		res.succeeded = moveCardBetweenZonesInBatch(r.game, card.Owner, ref.CardID,
			zone.Exile, zone.Library, prim.Bottom, 0) || res.succeeded
	}
	clearLinkedObjects(r.game, key)
	return res
}

func linkedExiledCard(g *game.Game, ref game.LinkedObjectRef) (*game.CardInstance, bool) {
	if ref.CardID == 0 {
		return nil, false
	}
	if ref.ObjectID != 0 {
		snapshot, ok := lastKnownObject(g, ref.ObjectID)
		if !ok || snapshot.CardID != ref.CardID {
			return nil, false
		}
	}
	card, ok := linkedCardInstance(g, ref)
	if !ok {
		return nil, false
	}
	owner, ok := playerByID(g, card.Owner)
	if !ok || !owner.Exile.Contains(card.ID) {
		return nil, false
	}
	return card, true
}
