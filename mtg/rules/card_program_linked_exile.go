package rules

import (
	"errors"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// validateExileDuration checks a placement's CR 607 linked publication and CR
// 610.3 "until this leaves the battlefield" duration before anything moves.
func (c *programContext) validateExileDuration(placement cardprogram.CardPlacement) error {
	if placement.PublishLinkedProducer == "" && !placement.UntilSourceLeaves {
		return nil
	}
	if placement.Destination != zone.Exile {
		return errors.New("linked or until-source-leaves placement requires exile")
	}
	if placement.PublishLinkedProducer != "" {
		if err := game.ValidateLinkedProducerOperation(
			game.MovePermanent{PublishLinkedProducer: placement.PublishLinkedProducer},
			c.obj.ResolvingLinkedProducers,
		); err != nil {
			return err
		}
		if originalLinkedProducerKey(c.game, c.obj, placement.PublishLinkedProducer).SourceID == 0 {
			return errors.New("linked producer requires an original battlefield source occurrence")
		}
	}
	if placement.UntilSourceLeaves {
		if c.obj.Kind != game.StackActivatedAbility && c.obj.Kind != game.StackTriggeredAbility {
			return errors.New("until-source-leaves exile requires an activated or triggered ability")
		}
		for _, member := range placement.Cards {
			if c.foundZones[member.Identity.CardID] != zone.Hand {
				return errors.New("until-source-leaves card placement requires hand handles")
			}
		}
	}
	return nil
}

// placementExileRecord accumulates the exact card incarnations one exile
// placement actually put into exile, for its linked publication and its
// until-source-leaves duration.
type placementExileRecord struct {
	publish game.LinkedObjectKey
	until   *untilSourceLeavesRecorder
}

// startPlacementExileRecord begins a validated placement's exile records. With
// UntilSourceLeaves, proceed is false when the source permanent already left
// the battlefield: CR 610.3a-b says the cards don't move.
func (c *programContext) startPlacementExileRecord(placement cardprogram.CardPlacement) (record *placementExileRecord, proceed bool) {
	if placement.PublishLinkedProducer == "" && !placement.UntilSourceLeaves {
		return nil, true
	}
	record = &placementExileRecord{}
	if placement.UntilSourceLeaves {
		sourceObjectID, present := untilSourceLeavesSource(c.game, c.obj)
		if !present {
			return nil, false
		}
		record.until = &untilSourceLeavesRecorder{sourceObjectID: sourceObjectID, returnZone: zone.Hand}
	}
	if placement.PublishLinkedProducer != "" {
		record.publish = originalLinkedProducerKey(c.game, c.obj, placement.PublishLinkedProducer)
	}
	return record, true
}

// add records card if its move actually put it into its owner's exile; a
// replacement effect can send it elsewhere instead. Only a face-up card is
// published for later linked abilities (CR 607.2a).
func (r *placementExileRecord) add(g *game.Game, card *game.CardInstance) {
	if r == nil || !g.Players[card.Owner].Exile.Contains(card.ID) {
		return
	}
	ref := game.LinkedObjectRef{CardID: card.ID, CardZoneVersion: card.ZoneVersion, CardZoneVersionSet: true}
	if r.until != nil {
		r.until.objects = append(r.until.objects, ref)
	}
	if r.publish.SourceID != 0 && !g.Players[card.Owner].Exile.IsFaceDown(card.ID) {
		rememberLinkedObject(g, r.publish, ref)
	}
}

func (r *placementExileRecord) commit(g *game.Game) {
	if r != nil {
		r.until.commit(g)
	}
}

func (c *programContext) LinkedExiledCards(producer game.LinkedProducerKey) ([]cardprogram.FoundCard, error) {
	if err := c.valid(); err != nil {
		return nil, err
	}
	if producer == "" {
		return nil, errors.New("linked exiled cards require a producer")
	}
	key := originalLinkedProducerKey(c.game, c.obj, producer)
	if key.SourceID == 0 {
		return nil, nil
	}
	var cards []cardprogram.FoundCard
	for _, ref := range linkedObjects(c.game, key) {
		card, ok := linkedCardInstance(c.game, ref)
		if !ok || !c.game.Players[card.Owner].Exile.Contains(card.ID) || c.game.Players[card.Owner].Exile.IsFaceDown(card.ID) {
			continue
		}
		cards = append(cards, c.mintFoundCard(card, zone.Exile))
	}
	return cards, nil
}
