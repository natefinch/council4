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
	for _, ref := range linkedObjects(r.game, key) {
		if snapshot, ok := lastKnownObject(r.game, ref.ObjectID); !ok || snapshot.CardID != ref.CardID {
			continue
		}
		card, ok := r.game.GetCardInstance(ref.CardID)
		if !ok {
			continue
		}
		owner, ok := playerByID(r.game, card.Owner)
		if !ok || !owner.Exile.Remove(ref.CardID) {
			continue
		}
		if prim.Bottom {
			owner.Library.AddToBottom(ref.CardID)
		} else {
			owner.Library.Add(ref.CardID)
		}
		emitZoneChangeEvent(r.game, game.Event{
			Player:   card.Owner,
			CardID:   ref.CardID,
			FromZone: zone.Exile,
			ToZone:   zone.Library,
		})
		res.succeeded = true
	}
	clearLinkedObjects(r.game, key)
	return res
}
