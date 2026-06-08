package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func createCardPermanent(g *game.Game, card *game.CardInstance, controller game.PlayerID, fromZone zone.Type) (*game.Permanent, bool) {
	return createCardPermanentFace(g, card, controller, fromZone, game.FaceFront)
}

func createCardPermanentWithChoices(e *Engine, g *game.Game, card *game.CardInstance, controller game.PlayerID, fromZone zone.Type, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (*game.Permanent, bool) {
	return createCardPermanentFaceWithChoices(e, g, card, controller, fromZone, game.FaceFront, agents, log)
}

func createCardPermanentFace(g *game.Game, card *game.CardInstance, controller game.PlayerID, fromZone zone.Type, face game.FaceIndex) (*game.Permanent, bool) {
	return createCardPermanentFaceWithChoices(NewEngine(nil), g, card, controller, fromZone, face, [game.NumPlayers]PlayerAgent{}, nil)
}

func createCardPermanentFaceWithChoices(e *Engine, g *game.Game, card *game.CardInstance, controller game.PlayerID, fromZone zone.Type, face game.FaceIndex, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (*game.Permanent, bool) {
	return createCardPermanentFaceWithContinuous(e, g, card, controller, fromZone, face, nil, agents, log)
}

func createCardPermanentFaceWithContinuous(e *Engine, g *game.Game, card *game.CardInstance, controller game.PlayerID, fromZone zone.Type, face game.FaceIndex, continuous []game.ContinuousEffect, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (*game.Permanent, bool) {
	return createCardPermanentFaceWithOptions(e, g, card, controller, fromZone, face, continuous, permanentCreationOptions{}, agents, log)
}

type permanentCreationOptions struct {
	ForceTapped bool
}

func createCardPermanentFaceWithOptions(e *Engine, g *game.Game, card *game.CardInstance, controller game.PlayerID, fromZone zone.Type, face game.FaceIndex, continuous []game.ContinuousEffect, options permanentCreationOptions, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (*game.Permanent, bool) {
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
	}
	initializePermanentCounters(permanent, faceDef)
	applyInitialContinuousEffects(g, permanent, continuous)
	applyEnterBattlefieldReplacementEffects(enterBattlefieldContext{
		engine: e,
		agents: agents,
		log:    log,
	}, g, permanent, fromZone)
	if options.ForceTapped {
		permanent.Tapped = true
	}
	g.Battlefield = append(g.Battlefield, permanent)
	event := game.Event{
		SourceID:    card.ID,
		Controller:  controller,
		Player:      card.Owner,
		CardID:      card.ID,
		Face:        face,
		PermanentID: objectID,
		FromZone:    fromZone,
		ToZone:      zone.Battlefield,
	}
	emitZoneChangeEvent(g, event)
	event.Kind = game.EventPermanentEnteredBattlefield
	emitEvent(g, event)
	return permanent, true
}

func applyInitialContinuousEffects(g *game.Game, permanent *game.Permanent, continuous []game.ContinuousEffect) {
	for i := range continuous {
		template := continuous[i]
		template.ID = g.IDGen.Next()
		template.SourceObjectID = permanent.ObjectID
		template.SourceCardID = permanent.CardInstanceID
		template.Controller = permanent.Controller
		template.Timestamp = permanent.Timestamp()
		template.AffectedObjectID = permanent.ObjectID
		if template.Duration == game.DurationPermanent {
			template.Duration = game.DurationPermanent
		}
		g.ContinuousEffects = append(g.ContinuousEffects, template)
	}
}

func createCardPermanentFaceDown(g *game.Game, card *game.CardInstance, controller game.PlayerID, fromZone zone.Type, face game.FaceIndex, kind game.FaceDownKind) (*game.Permanent, bool) {
	if _, ok := cardFaceDef(card, face); !ok || kind == game.FaceDownNone {
		return nil, false
	}
	objectID := g.IDGen.Next()
	permanent := &game.Permanent{
		ObjectID:       objectID,
		CardInstanceID: card.ID,
		Owner:          card.Owner,
		Controller:     controller,
		Face:           face,
		FaceDown:       true,
		FaceDownFace:   face,
		FaceDownKind:   kind,
		SummoningSick:  true,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	event := game.Event{
		SourceID:    card.ID,
		Controller:  controller,
		Player:      card.Owner,
		CardID:      card.ID,
		Face:        face,
		PermanentID: objectID,
		CardTypes:   []types.Card{types.Creature},
		FromZone:    fromZone,
		ToZone:      zone.Battlefield,
	}
	emitZoneChangeEvent(g, event)
	event.Kind = game.EventPermanentEnteredBattlefield
	emitEvent(g, event)
	return permanent, true
}

func initializePermanentCounters(permanent *game.Permanent, def *game.CardDef) {
	if def.HasSubtype(types.Class) {
		permanent.ClassLevel = 1
	}
	if def.Loyalty.Exists {
		permanent.Counters.Add(counter.Loyalty, def.Loyalty.Val)
	}
	if def.Defense.Exists {
		permanent.Counters.Add(counter.Defense, def.Defense.Val)
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

func movePermanentToZone(g *game.Game, permanent *game.Permanent, destination zone.Type) bool {
	if _, ok := permanentByObjectID(g, permanent.ObjectID); !ok {
		return false
	}
	snapshot := snapshotPermanent(g, permanent, zone.Battlefield)
	rememberLastKnown(g, &snapshot)
	event := game.Event{
		Kind:        game.EventZoneChanged,
		Controller:  effectiveController(g, permanent),
		Player:      permanent.Owner,
		CardID:      permanent.CardInstanceID,
		Face:        permanent.Face,
		PermanentID: permanent.ObjectID,
		TokenName:   permanentTokenName(permanent),
		TokenDef:    permanent.TokenDef,
		FromZone:    zone.Battlefield,
		ToZone:      destination,
	}
	actualDestination := replacementZoneChangeDestination(g, event)
	if !permanent.Token {
		actualDestination = commanderReplacementDestination(g, permanent.CardInstanceID, actualDestination)
	}
	if permanent.FaceDown {
		emitFaceDownRevealEvent(g, permanent)
	}
	detachPermanent(g, permanent)
	detachAttachmentsFromPermanent(g, permanent)
	removed, ok := removePermanentFromBattlefield(g, permanent.ObjectID)
	if !ok {
		return false
	}
	destinationCards, ok := destinationZone(g, removed.Owner, actualDestination)
	if !ok {
		return false
	}
	if removed.Token {
		destinationCards.Add(removed.ObjectID)
		emitPermanentLeaveEvents(g, removed, actualDestination)
		return true
	}

	destinationCards.Add(removed.CardInstanceID)
	emitPermanentLeaveEvents(g, removed, actualDestination)
	return true
}

func moveCardBetweenZones(g *game.Game, playerID game.PlayerID, cardID id.ID, fromZone, toZone zone.Type) bool {
	from, ok := destinationZone(g, playerID, fromZone)
	if !ok || !from.Remove(cardID) {
		return false
	}
	to, ok := destinationZone(g, playerID, toZone)
	if !ok {
		from.Add(cardID)
		return false
	}
	to.Add(cardID)
	emitZoneChangeEvent(g, game.Event{
		Player:   playerID,
		CardID:   cardID,
		FromZone: fromZone,
		ToZone:   toZone,
	})
	return true
}

func removeCardFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, fromZone zone.Type) bool {
	from, ok := destinationZone(g, playerID, fromZone)
	return ok && from.Remove(cardID)
}

func discardCardFromHand(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	player, ok := playerByID(g, playerID)
	if !ok || !player.Hand.Remove(cardID) {
		return false
	}
	card, cardOK := g.GetCardInstance(cardID)
	destination := zone.Graveyard
	if cardOK {
		if _, ok := madnessCostForCard(cardFaceOrDefault(card, game.FaceFront)); ok {
			destination = zone.Exile
		}
		destination = replacementZoneChangeDestination(g, game.Event{
			Kind:       game.EventZoneChanged,
			Controller: playerID,
			Player:     playerID,
			CardID:     cardID,
			FromZone:   zone.Hand,
			ToZone:     destination,
		})
		destination = commanderReplacementDestination(g, card.ID, destination)
	}
	zoneOwner := playerID
	if destination == zone.Command && cardOK {
		zoneOwner = card.Owner
	}
	destinationCards, ok := destinationZone(g, zoneOwner, destination)
	if !ok {
		return false
	}
	destinationCards.Add(cardID)
	event := game.Event{
		Player:   playerID,
		CardID:   cardID,
		FromZone: zone.Hand,
		ToZone:   destination,
		Amount:   1,
	}
	emitZoneChangeEvent(g, event)
	// A command-zone replacement changes the destination, but the discard still happened.
	event.Kind = game.EventCardDiscarded
	emitEvent(g, event)
	return true
}

func emitPermanentLeaveEvents(g *game.Game, permanent *game.Permanent, destination zone.Type) {
	event := game.Event{
		Controller:  permanent.Controller,
		Player:      permanent.Owner,
		CardID:      permanent.CardInstanceID,
		Face:        permanent.Face,
		PermanentID: permanent.ObjectID,
		TokenName:   permanentTokenName(permanent),
		TokenDef:    permanent.TokenDef,
		FromZone:    zone.Battlefield,
		ToZone:      destination,
	}
	emitZoneChangeEvent(g, event)
	if destination == zone.Graveyard {
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
	if commanderReplacementDestination(g, permanent.CardInstanceID, zone.Graveyard) == zone.Command {
		movePermanentToZone(g, permanent, zone.Graveyard)
		return nil, false
	}
	if !movePermanentToZone(g, permanent, zone.Graveyard) {
		return nil, false
	}
	return permanent, true
}

func destinationZone(g *game.Game, owner game.PlayerID, destination zone.Type) (*zone.Zone, bool) {
	if owner < 0 || int(owner) >= len(g.Players) {
		return nil, false
	}
	player := g.Players[owner]
	switch destination {
	case zone.Library:
		return &player.Library, true
	case zone.Hand:
		return &player.Hand, true
	case zone.Graveyard:
		return &player.Graveyard, true
	case zone.Exile:
		return &player.Exile, true
	case zone.Command:
		return &player.CommandZone, true
	default:
		return nil, false
	}
}
