package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func (r *effectResolver) placeObservedLibraryCard(card *game.CardInstance, bottom bool) effectResolved {
	result := effectResolved{accepted: true}
	library, ok := destinationZone(r.game, card.Owner, zone.Library)
	if !ok || !library.Remove(card.ID) {
		return result
	}
	if bottom {
		library.AddToBottom(card.ID)
	} else {
		library.Add(card.ID)
	}
	r.rememberResultCard(card.ID)
	result.succeeded = true
	result.amount = 1
	return result
}
