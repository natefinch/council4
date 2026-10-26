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
// identity, so an object that already left exile is skipped. Every card, across
// all owners, moves as one simultaneous batch (CR 603.2c): each card's
// zone-change replacements and its owner's choices are determined in APNAP order
// before any card moves (CR 101.4, 400.6, 616.1). RandomOrder randomizes each
// owner's cards independently.
func handlePutLinkedExiledCardsInLibrary(r *effectResolver, prim game.PutLinkedExiledCardsInLibrary) effectResolved {
	res := effectResolved{accepted: true}
	key := linkedObjectSourceKey(r.game, r.obj, string(prim.LinkedKey))
	var owners []game.PlayerID
	cardsByOwner := make(map[game.PlayerID][]game.ZoneCardSnapshot)
	for _, ref := range linkedObjects(r.game, key) {
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
	var moves []simultaneousCardMove
	for _, owner := range owners {
		cards := cardsByOwner[owner]
		if prim.RandomOrder {
			cards = randomlyOrderedSnapshots(cards, r.engine.rng)
		}
		for _, snapshot := range cards {
			moves = append(moves, simultaneousCardMove{card: snapshot, fromZone: zone.Exile})
		}
	}
	if len(moves) > 0 {
		res.succeeded = moveCardsBetweenZonesSimultaneously(r.game, moves, zone.Library, prim.Bottom, r.game.IDGen.Next())
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
