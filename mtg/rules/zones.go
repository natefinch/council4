package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
)

func createCardPermanent(g *game.Game, card *game.CardInstance, controller game.PlayerID, fromZone game.ZoneType) (*game.Permanent, bool) {
	return createCardPermanentFace(g, card, controller, fromZone, game.FaceFront)
}

func createCardPermanentFace(g *game.Game, card *game.CardInstance, controller game.PlayerID, fromZone game.ZoneType, face game.FaceIndex) (*game.Permanent, bool) {
	faceDef, ok := cardFaceDef(card, face)
	if !ok {
		return nil, false
	}
	objectID := g.IDGen.Next()
	permanent := &game.Permanent{
		ObjectID:       objectID,
		CardInstanceID: card.ID,
		Owner:          card.Owner,
		Controller:     controller,
		Face:           face,
		SummoningSick:  entersSummoningSick(faceDef),
		Timestamp:      int64(objectID),
	}
	initializePermanentCounters(permanent, faceDef)
	applyEnterBattlefieldReplacementEffects(g, permanent, fromZone)
	g.Battlefield = append(g.Battlefield, permanent)
	event := game.GameEvent{
		SourceID:    card.ID,
		Controller:  controller,
		Player:      card.Owner,
		CardID:      card.ID,
		Face:        face,
		PermanentID: objectID,
		FromZone:    fromZone,
		ToZone:      game.ZoneBattlefield,
	}
	emitZoneChangeEvent(g, event)
	event.Kind = game.EventPermanentEnteredBattlefield
	emitEvent(g, event)
	return permanent, true
}

func initializePermanentCounters(permanent *game.Permanent, def *game.CardDef) {
	if def.EntersTapped {
		permanent.Tapped = true
	}
	if def.Loyalty.Exists {
		permanent.Counters.Add(counter.Loyalty, def.Loyalty.Val)
	}
	if def.Defense.Exists {
		permanent.Counters.Add(counter.Defense, def.Defense.Val)
	}
	for _, placement := range def.EntersWithCounters {
		permanent.Counters.Add(placement.Kind, placement.Amount)
	}
}

func removePermanentFromBattlefield(g *game.Game, objectID id.ID) (*game.Permanent, bool) {
	for i, permanent := range g.Battlefield {
		if permanent.ObjectID != objectID {
			continue
		}
		g.Battlefield = append(g.Battlefield[:i], g.Battlefield[i+1:]...)
		return permanent, true
	}
	return nil, false
}

func movePermanentToZone(g *game.Game, permanent *game.Permanent, destination game.ZoneType) bool {
	if _, ok := permanentByObjectID(g, permanent.ObjectID); !ok {
		return false
	}
	rememberLastKnown(g, snapshotPermanent(g, permanent, game.ZoneBattlefield))
	event := game.GameEvent{
		Kind:        game.EventZoneChanged,
		Controller:  effectiveController(g, permanent),
		Player:      permanent.Owner,
		CardID:      permanent.CardInstanceID,
		Face:        permanent.Face,
		PermanentID: permanent.ObjectID,
		TokenName:   permanentTokenName(permanent),
		TokenDef:    permanent.TokenDef,
		FromZone:    game.ZoneBattlefield,
		ToZone:      destination,
	}
	actualDestination := replacementZoneChangeDestination(g, event)
	if !permanent.Token {
		actualDestination = commanderReplacementDestination(g, permanent.CardInstanceID, actualDestination)
	}
	detachPermanent(g, permanent)
	detachAttachmentsFromPermanent(g, permanent)
	removed, ok := removePermanentFromBattlefield(g, permanent.ObjectID)
	if !ok {
		return false
	}
	zone, ok := destinationZone(g, removed.Owner, actualDestination)
	if !ok {
		return false
	}
	if removed.Token {
		zone.Add(removed.ObjectID)
		emitPermanentLeaveEvents(g, removed, actualDestination)
		return true
	}

	zone.Add(removed.CardInstanceID)
	emitPermanentLeaveEvents(g, removed, actualDestination)
	return true
}

func discardCardFromHand(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	player, ok := playerByID(g, playerID)
	if !ok || !player.Hand.Remove(cardID) {
		return false
	}
	card, cardOK := g.GetCardInstance(cardID)
	destination := game.ZoneGraveyard
	if cardOK {
		destination = replacementZoneChangeDestination(g, game.GameEvent{
			Kind:       game.EventZoneChanged,
			Controller: playerID,
			Player:     playerID,
			CardID:     cardID,
			FromZone:   game.ZoneHand,
			ToZone:     destination,
		})
		destination = commanderReplacementDestination(g, card.ID, destination)
	}
	zoneOwner := playerID
	if destination == game.ZoneCommand && cardOK {
		zoneOwner = card.Owner
	}
	zone, ok := destinationZone(g, zoneOwner, destination)
	if !ok {
		return false
	}
	zone.Add(cardID)
	event := game.GameEvent{
		Player:   playerID,
		CardID:   cardID,
		FromZone: game.ZoneHand,
		ToZone:   destination,
		Amount:   1,
	}
	emitZoneChangeEvent(g, event)
	// A command-zone replacement changes the destination, but the discard still happened.
	event.Kind = game.EventCardDiscarded
	emitEvent(g, event)
	return true
}

func emitPermanentLeaveEvents(g *game.Game, permanent *game.Permanent, destination game.ZoneType) {
	event := game.GameEvent{
		Controller:  permanent.Controller,
		Player:      permanent.Owner,
		CardID:      permanent.CardInstanceID,
		Face:        permanent.Face,
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
	permanent, ok := permanentByObjectID(g, objectID)
	if !ok {
		return nil, false
	}
	if hasKeyword(g, permanent, game.Indestructible) {
		return nil, false
	}
	if replaceDestroyPermanent(g, permanent) {
		return nil, false
	}
	if commanderReplacementDestination(g, permanent.CardInstanceID, game.ZoneGraveyard) == game.ZoneCommand {
		movePermanentToZone(g, permanent, game.ZoneGraveyard)
		return nil, false
	}
	if !movePermanentToZone(g, permanent, game.ZoneGraveyard) {
		return nil, false
	}
	return permanent, true
}

func destinationZone(g *game.Game, owner game.PlayerID, destination game.ZoneType) (*game.Zone, bool) {
	if owner < 0 || int(owner) >= len(g.Players) {
		return nil, false
	}
	player := g.Players[owner]
	switch destination {
	case game.ZoneLibrary:
		return &player.Library, true
	case game.ZoneHand:
		return &player.Hand, true
	case game.ZoneGraveyard:
		return &player.Graveyard, true
	case game.ZoneExile:
		return &player.Exile, true
	case game.ZoneCommand:
		return &player.CommandZone, true
	default:
		return nil, false
	}
}
