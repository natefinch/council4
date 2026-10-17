package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

func originalLinkedProducerKey(g *game.Game, obj *game.StackObject, producer game.LinkedProducerKey) game.LinkedObjectKey {
	if obj == nil || obj.Kind == game.StackSpell {
		return game.LinkedObjectKey{}
	}
	// SourceID is frozen when an ability is queued; never look up a replacement
	// permanent by physical CardID after the original source leaves.
	if _, live := permanentByObjectID(g, obj.SourceID); !live {
		if snapshot, known := lastKnownObject(g, obj.SourceID); !known || snapshot.FromZone != zone.Battlefield {
			return game.LinkedObjectKey{}
		}
	}
	return game.LinkedObjectKey{SourceID: obj.SourceID, LinkID: string(producer), Producer: producer}
}

func (r *effectResolver) originalProducerPublicationKey(producer game.LinkedProducerKey) game.LinkedObjectKey {
	if r.obj == nil {
		panic(UnsupportedError{Reason: "linked producer requires a resolving ability"})
	}
	if err := game.ValidateLinkedProducerOperation(game.MovePermanent{PublishLinkedProducer: producer}, r.obj.ResolvingLinkedProducers); err != nil {
		panic(UnsupportedError{Reason: err.Error()})
	}
	key := originalLinkedProducerKey(r.game, r.obj, producer)
	if key.SourceID == 0 {
		panic(UnsupportedError{Reason: "linked producer requires an original battlefield source occurrence"})
	}
	return key
}

func (r *effectResolver) publishProducerPermanent(key game.LinkedObjectKey, permanent *game.Permanent) {
	snapshot, ok := lastKnownObject(r.game, permanent.ObjectID)
	if !ok {
		return
	}
	for _, zoneCard := range snapshot.ZoneCards {
		card, ok := r.game.GetCardInstance(zoneCard.CardID)
		if !ok || card.ZoneVersion != zoneCard.ZoneVersion || !r.game.Players[card.Owner].Exile.Contains(card.ID) {
			continue
		}
		rememberLinkedObject(r.game, key, game.LinkedObjectRef{
			ObjectID: snapshot.ObjectID, CardID: card.ID, CardZoneVersion: card.ZoneVersion, CardZoneVersionSet: true,
		})
	}
}

func (r *effectResolver) returnOriginalProducerCards(prim game.PutOnBattlefield) []*game.Permanent {
	sources := prim.Sources
	if prim.Source.Valid() {
		sources = []game.BattlefieldSource{prim.Source}
	}
	var resolved []resolvedBattlefieldCard
	seen := make(map[id.ID]bool)
	for _, source := range sources {
		producer, ok := source.LinkedProducer()
		if !ok {
			return nil
		}
		key := originalLinkedProducerKey(r.game, r.obj, producer)
		for _, ref := range linkedObjects(r.game, key) {
			card, ok := linkedCardInstance(r.game, ref)
			if !ok || seen[card.ID] || !r.game.Players[card.Owner].Exile.Contains(card.ID) ||
				!card.Def.IsPermanent() || !cardMatchesSelection(r.game, r.obj, card, prim.Selection) {
				continue
			}
			controller := card.Owner
			if prim.Recipient.Exists {
				var valid bool
				controller, valid = r.recipientController(prim.Recipient.Val)
				if !valid {
					return nil
				}
			}
			seen[card.ID] = true
			resolved = append(resolved, resolvedBattlefieldCard{card: card, fromZone: zone.Exile, controller: controller})
		}
	}
	returned := r.putResolvedCardsOnBattlefieldCollecting(resolved, prim.ContinuousEffects, battlefieldEntryOptions(prim))
	// CR 607 links are not one-shot CR 610.3 captures: an inadmissible Aura
	// remains linked in exile for a later resolving copy. Exact incarnations
	// make successful or redirected moves unavailable without consuming others.
	return returned
}
