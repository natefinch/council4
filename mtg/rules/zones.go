package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
)

func createCardPermanent(g *game.Game, card *game.CardInstance, controller game.PlayerID, fromZone game.ZoneType) *game.Permanent {
	if g == nil || card == nil || card.Def == nil {
		return nil
	}
	objectID := g.IDGen.Next()
	permanent := &game.Permanent{
		ObjectID:       objectID,
		CardInstanceID: card.ID,
		Owner:          card.Owner,
		Controller:     controller,
		SummoningSick:  entersSummoningSick(card.Def),
		Timestamp:      int64(objectID),
	}
	initializePermanentCounters(permanent, card.Def)
	g.Battlefield = append(g.Battlefield, permanent)
	event := game.GameEvent{
		SourceID:    card.ID,
		Controller:  controller,
		Player:      card.Owner,
		CardID:      card.ID,
		PermanentID: objectID,
		FromZone:    fromZone,
		ToZone:      game.ZoneBattlefield,
	}
	emitZoneChangeEvent(g, event)
	event.Kind = game.EventPermanentEnteredBattlefield
	emitEvent(g, event)
	return permanent
}

func initializePermanentCounters(permanent *game.Permanent, def *game.CardDef) {
	if permanent == nil || def == nil {
		return
	}
	if def.Loyalty != nil {
		permanent.Counters.Add(counter.Loyalty, *def.Loyalty)
	}
	if def.Defense != nil {
		permanent.Counters.Add(counter.Defense, *def.Defense)
	}
}

func removePermanentFromBattlefield(g *game.Game, objectID id.ID) *game.Permanent {
	if g == nil {
		return nil
	}
	for i, permanent := range g.Battlefield {
		if permanent == nil || permanent.ObjectID != objectID {
			continue
		}
		g.Battlefield = append(g.Battlefield[:i], g.Battlefield[i+1:]...)
		return permanent
	}
	return nil
}

func movePermanentToZone(g *game.Game, permanent *game.Permanent, destination game.ZoneType) bool {
	if g == nil || permanent == nil {
		return false
	}

	detachPermanent(g, permanent)
	detachAttachmentsFromPermanent(g, permanent)
	removed := removePermanentFromBattlefield(g, permanent.ObjectID)
	if removed == nil {
		return false
	}
	zone := destinationZone(g, removed.Owner, destination)
	if zone == nil {
		return false
	}
	if removed.Token {
		zone.Add(removed.ObjectID)
		emitPermanentLeaveEvents(g, removed, destination)
		return true
	}

	zone.Add(removed.CardInstanceID)
	emitPermanentLeaveEvents(g, removed, destination)
	return true
}

func discardCardFromHand(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	player := playerByID(g, playerID)
	if player == nil || !player.Hand.Remove(cardID) {
		return false
	}
	player.Graveyard.Add(cardID)
	event := game.GameEvent{
		Player:   playerID,
		CardID:   cardID,
		FromZone: game.ZoneHand,
		ToZone:   game.ZoneGraveyard,
		Amount:   1,
	}
	emitZoneChangeEvent(g, event)
	event.Kind = game.EventCardDiscarded
	emitEvent(g, event)
	return true
}

func emitPermanentLeaveEvents(g *game.Game, permanent *game.Permanent, destination game.ZoneType) {
	if permanent == nil {
		return
	}
	event := game.GameEvent{
		Controller:  permanent.Controller,
		Player:      permanent.Owner,
		CardID:      permanent.CardInstanceID,
		PermanentID: permanent.ObjectID,
		TokenName:   permanentTokenName(permanent),
		TokenDef:    permanent.TokenDef,
		FromZone:    game.ZoneBattlefield,
		ToZone:      destination,
	}
	emitZoneChangeEvent(g, event)
	if destination == game.ZoneGraveyard {
		event.Kind = game.EventPermanentDied
		emitEvent(g, event)
	}
}

func destroyPermanent(g *game.Game, objectID id.ID) (*game.Permanent, bool) {
	permanent := permanentByObjectID(g, objectID)
	if permanent == nil {
		return nil, false
	}
	if hasKeyword(g, permanent, game.Indestructible) {
		return nil, false
	}
	if replaceDestroyPermanent(g, permanent) {
		return nil, false
	}
	if !movePermanentToZone(g, permanent, game.ZoneGraveyard) {
		return nil, false
	}
	return permanent, true
}

func destinationZone(g *game.Game, owner game.PlayerID, destination game.ZoneType) *game.Zone {
	if g == nil || owner < 0 || int(owner) >= len(g.Players) {
		return nil
	}
	player := g.Players[owner]
	if player == nil {
		return nil
	}
	switch destination {
	case game.ZoneLibrary:
		return &player.Library
	case game.ZoneHand:
		return &player.Hand
	case game.ZoneGraveyard:
		return &player.Graveyard
	case game.ZoneExile:
		return &player.Exile
	case game.ZoneCommand:
		return &player.CommandZone
	default:
		return nil
	}
}
