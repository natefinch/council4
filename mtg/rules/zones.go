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
	ForceTapped       bool
	KickerPaid        bool
	Evoked            bool
	WasCast           bool
	CastController    game.PlayerID
	HasCastController bool
	CastFromZone      zone.Type
	Counters          []game.CounterPlacement
	SimultaneousID    id.ID
	XValue            int
	// ColorsOfManaSpentToCast carries the number of distinct colors of mana
	// spent to cast the spell that is resolving into this permanent, so a
	// Converge enters-with-counters replacement ("for each color of mana spent
	// to cast it") reads the count as the permanent enters. It is zero for a
	// permanent that did not enter from a cast spell (a token, a copy, a
	// put-into-play effect).
	ColorsOfManaSpentToCast int
}

func createCardPermanentFaceWithOptions(e *Engine, g *game.Game, card *game.CardInstance, controller game.PlayerID, fromZone zone.Type, face game.FaceIndex, continuous []game.ContinuousEffect, options permanentCreationOptions, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (*game.Permanent, bool) {
	faceDef, ok := cardFaceDef(card, face)
	if !ok {
		return nil, false
	}
	if entryFromZoneProhibited(g, faceDef, fromZone) {
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
		Prepared:       faceDef.EntersPrepared,
	}
	initializePermanentCounters(permanent, faceDef)
	applyInitialContinuousEffects(g, permanent, continuous)
	registerPermanentReplacementEffects(g, permanent)
	initializeReadAhead(e, g, permanent, agents, log)
	if optionalEntryReplacementDeclined(enterBattlefieldContext{
		engine: e,
		agents: agents,
		log:    log,
	}, g, card, permanent, faceDef, fromZone) {
		return nil, false
	}
	applyEnterBattlefieldReplacementEffects(enterBattlefieldContext{
		engine:            e,
		agents:            agents,
		log:               log,
		xValue:            options.XValue,
		colorsOfManaSpent: options.ColorsOfManaSpentToCast,
	}, g, permanent, fromZone)
	if options.ForceTapped {
		permanent.Tapped = true
	}
	for _, placement := range options.Counters {
		permanent.Counters.Add(placement.Kind, placement.Amount)
	}
	g.Battlefield = append(g.Battlefield, permanent)
	if lore := permanent.Counters.Get(counter.Lore); lore > 0 {
		emitCounterAddedEvent(g, permanent, effectiveController(g, permanent), counter.Lore, 0, lore)
	}
	event := game.Event{
		SourceID:               card.ID,
		Controller:             controller,
		Player:                 card.Owner,
		CardID:                 card.ID,
		Face:                   face,
		KickerPaid:             options.KickerPaid,
		EnterEvoked:            options.Evoked,
		EnterWasCast:           options.WasCast,
		EnterCastController:    options.CastController,
		EnterHasCastController: options.HasCastController,
		EnterCastFromZone:      options.CastFromZone,
		PermanentID:            objectID,
		FromZone:               fromZone,
		ToZone:                 zone.Battlefield,
		SimultaneousID:         options.SimultaneousID,
	}
	event = emitZoneChangeEvent(g, event)
	event.Kind = game.EventPermanentEnteredBattlefield
	emitEvent(g, event)
	return permanent, true
}

type preparedCardPermanentEntry struct {
	card       *game.CardInstance
	permanent  *game.Permanent
	controller game.PlayerID
	fromZone   zone.Type
	continuous []game.ContinuousEffect
	options    permanentCreationOptions
}

func prepareCardPermanentFaceForSimultaneousEntry(
	e *Engine,
	g *game.Game,
	card *game.CardInstance,
	controller game.PlayerID,
	fromZone zone.Type,
	face game.FaceIndex,
	continuous []game.ContinuousEffect,
	options permanentCreationOptions,
	agents [game.NumPlayers]PlayerAgent,
	log *TurnLog,
) (preparedCardPermanentEntry, bool) {
	faceDef, ok := cardFaceDef(card, face)
	if !ok {
		return preparedCardPermanentEntry{}, false
	}
	if entryFromZoneProhibited(g, faceDef, fromZone) {
		return preparedCardPermanentEntry{}, false
	}
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: card.ID,
		Owner:          card.Owner,
		Controller:     controller,
		Face:           face,
		SummoningSick:  entersSummoningSick(faceDef),
		Prepared:       faceDef.EntersPrepared,
	}
	initializePermanentCounters(permanent, faceDef)
	initializeReadAhead(e, g, permanent, agents, log)
	applyEnterBattlefieldReplacementEffects(enterBattlefieldContext{
		engine:            e,
		agents:            agents,
		log:               log,
		xValue:            options.XValue,
		colorsOfManaSpent: options.ColorsOfManaSpentToCast,
	}, g, permanent, fromZone)
	if options.ForceTapped {
		permanent.Tapped = true
	}
	for _, placement := range options.Counters {
		permanent.Counters.Add(placement.Kind, placement.Amount)
	}
	return preparedCardPermanentEntry{
		card:       card,
		permanent:  permanent,
		controller: controller,
		fromZone:   fromZone,
		continuous: continuous,
		options:    options,
	}, true
}

func commitSimultaneousCardPermanentEntries(g *game.Game, entries []preparedCardPermanentEntry) {
	for i := range entries {
		entry := &entries[i]
		applyInitialContinuousEffects(g, entry.permanent, entry.continuous)
		registerPermanentReplacementEffects(g, entry.permanent)
	}
	for i := range entries {
		g.Battlefield = append(g.Battlefield, entries[i].permanent)
	}
	for i := range entries {
		entry := &entries[i]
		permanent := entry.permanent
		if lore := permanent.Counters.Get(counter.Lore); lore > 0 {
			emitCounterAddedEvent(g, permanent, effectiveController(g, permanent), counter.Lore, 0, lore)
		}
		event := game.Event{
			SourceID:               entry.card.ID,
			Controller:             entry.controller,
			Player:                 entry.card.Owner,
			CardID:                 entry.card.ID,
			Face:                   permanent.Face,
			KickerPaid:             entry.options.KickerPaid,
			EnterEvoked:            entry.options.Evoked,
			EnterWasCast:           entry.options.WasCast,
			EnterCastController:    entry.options.CastController,
			EnterHasCastController: entry.options.HasCastController,
			EnterCastFromZone:      entry.options.CastFromZone,
			PermanentID:            permanent.ObjectID,
			FromZone:               entry.fromZone,
			ToZone:                 zone.Battlefield,
			SimultaneousID:         entry.options.SimultaneousID,
		}
		event = emitZoneChangeEvent(g, event)
		event.Kind = game.EventPermanentEnteredBattlefield
		emitEvent(g, event)
	}
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

func createCardPermanentFaceDown(g *game.Game, card *game.CardInstance, controller game.PlayerID, fromZone zone.Type, face game.FaceIndex, kind game.FaceDownKind, wasCast bool) (*game.Permanent, bool) {
	return createCardPermanentFaceDownWithChoices(NewEngine(nil), g, card, controller, fromZone, face, kind, wasCast, [game.NumPlayers]PlayerAgent{}, nil)
}

func createCardPermanentFaceDownWithChoices(e *Engine, g *game.Game, card *game.CardInstance, controller game.PlayerID, fromZone zone.Type, face game.FaceIndex, kind game.FaceDownKind, wasCast bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (*game.Permanent, bool) {
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
	applyEnterBattlefieldReplacementEffects(enterBattlefieldContext{
		engine: e,
		agents: agents,
		log:    log,
	}, g, permanent, fromZone)
	g.Battlefield = append(g.Battlefield, permanent)
	event := game.Event{
		SourceID:               card.ID,
		Controller:             controller,
		Player:                 card.Owner,
		CardID:                 card.ID,
		Face:                   face,
		FaceDown:               true,
		EnterWasCast:           wasCast,
		EnterCastController:    controller,
		EnterHasCastController: wasCast,
		PermanentID:            objectID,
		CardTypes:              []types.Card{types.Creature},
		FromZone:               fromZone,
		ToZone:                 zone.Battlefield,
	}
	event = emitZoneChangeEvent(g, event)
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
	if def.HasSubtype(types.Saga) {
		permanent.Counters.Add(counter.Lore, 1)
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

type preparedPermanentZoneMove struct {
	permanent         *game.Permanent
	snapshot          game.ObjectSnapshot
	event             game.Event
	replacement       zoneChangeReplacementResult
	actualDestination zone.Type
	componentMoves    []mergedCardZoneMove
}

func preparePermanentZoneMove(g *game.Game, permanent *game.Permanent, destination zone.Type) (preparedPermanentZoneMove, bool) {
	if _, ok := permanentByObjectID(g, permanent.ObjectID); !ok {
		return preparedPermanentZoneMove{}, false
	}
	snapshot := snapshotPermanent(g, permanent, zone.Battlefield)
	event := game.Event{
		Kind:        game.EventZoneChanged,
		Controller:  effectiveController(g, permanent),
		Player:      permanent.Owner,
		CardID:      permanent.CardInstanceID,
		Face:        permanent.Face,
		FaceDown:    permanent.FaceDown,
		PermanentID: permanent.ObjectID,
		TokenName:   permanentTokenName(permanent),
		TokenDef:    permanent.TokenDef,
		FromZone:    zone.Battlefield,
		ToZone:      destination,
	}
	replacement := replacementZoneChange(g, event)
	replacedDestination := replacement.destination
	actualDestination := replacedDestination
	if !permanent.Token {
		actualDestination = commanderReplacementDestination(g, permanent.CardInstanceID, actualDestination)
	}
	componentMoves, ok := mergedComponentZoneMoves(g, permanent, replacedDestination)
	if !ok {
		return preparedPermanentZoneMove{}, false
	}
	if _, ok := destinationZone(g, permanent.Owner, actualDestination); !ok {
		return preparedPermanentZoneMove{}, false
	}
	return preparedPermanentZoneMove{
		permanent:         permanent,
		snapshot:          snapshot,
		event:             event,
		replacement:       replacement,
		actualDestination: actualDestination,
		componentMoves:    componentMoves,
	}, true
}

func applyPreparedPermanentZoneMove(g *game.Game, move *preparedPermanentZoneMove) bool {
	rememberLastKnown(g, &move.snapshot)
	revealZoneReplacementSource(g, move.event, move.replacement.revealSource)
	if move.permanent.FaceDown {
		emitFaceDownRevealEvent(g, move.permanent)
	}
	detachPermanent(g, move.permanent)
	detachAttachmentsFromPermanent(g, move.permanent)
	removed, ok := removePermanentFromBattlefield(g, move.permanent.ObjectID)
	if !ok {
		return false
	}
	destinationCards, _ := destinationZone(g, removed.Owner, move.actualDestination)
	if removed.Token {
		destinationCards.Add(removed.ObjectID)
		emitPermanentLeaveEvents(g, removed, move.event.Controller, move.actualDestination, move.event.SimultaneousID)
	} else {
		destinationCards.Add(removed.CardInstanceID)
		shuffleLibraryIfRequested(g, destinationCards, move.actualDestination, move.replacement.shuffleIntoLibrary)
		emitPermanentLeaveEvents(g, removed, move.event.Controller, move.actualDestination, move.event.SimultaneousID)
	}
	for _, component := range move.componentMoves {
		if component.faceDown {
			emitEvent(g, game.Event{
				Kind:       game.EventCardRevealed,
				Controller: move.event.Controller,
				Player:     component.owner,
				CardID:     component.cardID,
				Face:       component.faceDownFace,
				TokenName:  permanentTokenDefName(component.tokenDef),
				TokenDef:   component.tokenDef,
			})
		}
		if component.tokenDef != nil {
			emitZoneChangeEvent(g, game.Event{
				Controller:     move.event.Controller,
				Player:         component.owner,
				Face:           component.face,
				TokenDef:       component.tokenDef,
				TokenName:      component.tokenDef.Name,
				FromZone:       zone.Battlefield,
				ToZone:         component.destination,
				SimultaneousID: move.event.SimultaneousID,
			})
			continue
		}
		cards, ok := destinationZone(g, component.owner, component.destination)
		if !ok {
			panic("validated merged-card destination disappeared")
		}
		cards.Add(component.cardID)
		emitZoneChangeEvent(g, game.Event{
			Controller:     move.event.Controller,
			Player:         component.owner,
			CardID:         component.cardID,
			Face:           component.face,
			FromZone:       zone.Battlefield,
			ToZone:         component.destination,
			SimultaneousID: move.event.SimultaneousID,
		})
	}
	return true
}

func movePermanentToZone(g *game.Game, permanent *game.Permanent, destination zone.Type) bool {
	return movePermanentToZoneInBatch(g, permanent, destination, 0)
}

func movePermanentToZoneInBatch(g *game.Game, permanent *game.Permanent, destination zone.Type, simultaneousID id.ID) bool {
	move, ok := preparePermanentZoneMove(g, permanent, destination)
	if !ok {
		return false
	}
	move.event.SimultaneousID = simultaneousID
	return applyPreparedPermanentZoneMove(g, &move)
}

func movePermanentsToZoneSimultaneously(g *game.Game, permanents []*game.Permanent, destination zone.Type) bool {
	moves := make([]preparedPermanentZoneMove, 0, len(permanents))
	for _, permanent := range permanents {
		move, ok := preparePermanentZoneMove(g, permanent, destination)
		if ok {
			moves = append(moves, move)
		}
	}
	if len(moves) > 0 {
		simultaneousID := g.IDGen.Next()
		for i := range moves {
			moves[i].event.SimultaneousID = simultaneousID
		}
	}
	succeeded := false
	for i := range moves {
		succeeded = applyPreparedPermanentZoneMove(g, &moves[i]) || succeeded
	}
	return succeeded
}

type mergedCardZoneMove struct {
	cardID       id.ID
	face         game.FaceIndex
	faceDown     bool
	faceDownFace game.FaceIndex
	owner        game.PlayerID
	destination  zone.Type
	tokenDef     *game.CardDef
}

func mergedComponentZoneMoves(g *game.Game, permanent *game.Permanent, destination zone.Type) ([]mergedCardZoneMove, bool) {
	moves := make([]mergedCardZoneMove, 0, len(permanent.MergedCards))
	for _, component := range permanent.MergedCards {
		if component.TokenDef != nil {
			moves = append(moves, mergedCardZoneMove{
				face:         component.Face,
				faceDown:     component.FaceDown,
				faceDownFace: component.FaceDownFace,
				owner:        component.Owner,
				destination:  destination,
				tokenDef:     component.TokenDef,
			})
			continue
		}
		card, ok := g.GetCardInstance(component.CardInstanceID)
		if !ok {
			return nil, false
		}
		actualDestination := commanderReplacementDestination(g, card.ID, destination)
		if _, ok := destinationZone(g, card.Owner, actualDestination); !ok {
			return nil, false
		}
		moves = append(moves, mergedCardZoneMove{
			cardID:       card.ID,
			face:         component.Face,
			faceDown:     component.FaceDown,
			faceDownFace: component.FaceDownFace,
			owner:        card.Owner,
			destination:  actualDestination,
		})
	}
	return moves, true
}

func permanentTokenDefName(def *game.CardDef) string {
	if def == nil {
		return ""
	}
	return def.Name
}

func moveCardBetweenZones(g *game.Game, playerID game.PlayerID, cardID id.ID, fromZone, toZone zone.Type) bool {
	return moveCardBetweenZonesWithPlacement(g, playerID, cardID, fromZone, toZone, false)
}

func moveCardBetweenZonesWithPlacement(g *game.Game, playerID game.PlayerID, cardID id.ID, fromZone, toZone zone.Type, bottom bool) bool {
	return moveCardBetweenZonesInBatch(g, playerID, cardID, fromZone, toZone, bottom, 0)
}

func moveCardBetweenZonesInBatch(g *game.Game, playerID game.PlayerID, cardID id.ID, fromZone, toZone zone.Type, bottom bool, simultaneousID id.ID) bool {
	replacement := zoneChangeReplacementResult{destination: toZone}
	card, cardOK := g.GetCardInstance(cardID)
	event := game.Event{}
	if cardOK {
		event = game.Event{
			Kind:       game.EventZoneChanged,
			Controller: playerID,
			Player:     playerID,
			CardID:     cardID,
			FromZone:   fromZone,
			ToZone:     toZone,
		}
		replacement = replacementZoneChange(g, event)
		destination := replacement.destination
		destination = commanderReplacementDestination(g, card.ID, destination)
		replacement.destination = destination
	}
	return moveCardBetweenZonesAfterReplacement(g, playerID, cardID, fromZone, replacement, event, bottom, simultaneousID)
}

func moveCardBetweenZonesAfterReplacement(
	g *game.Game,
	playerID game.PlayerID,
	cardID id.ID,
	fromZone zone.Type,
	replacement zoneChangeReplacementResult,
	event game.Event,
	bottom bool,
	simultaneousID id.ID,
) bool {
	destination := replacement.destination
	from, ok := destinationZone(g, playerID, fromZone)
	if !ok || !from.Remove(cardID) {
		return false
	}
	zoneOwner := playerID
	if destination == zone.Command {
		if card, ok := g.GetCardInstance(cardID); ok {
			zoneOwner = card.Owner
		}
	}
	to, ok := destinationZone(g, zoneOwner, destination)
	if !ok {
		from.Add(cardID)
		return false
	}
	revealZoneReplacementSource(g, event, replacement.revealSource)
	if bottom && destination == zone.Library {
		to.AddToBottom(cardID)
	} else {
		to.Add(cardID)
	}
	shuffleLibraryIfRequested(g, to, destination, replacement.shuffleIntoLibrary)
	emitZoneChangeEvent(g, game.Event{
		Player:         playerID,
		CardID:         cardID,
		FromZone:       fromZone,
		ToZone:         destination,
		SimultaneousID: simultaneousID,
	})
	return true
}

func removeCardFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, fromZone zone.Type) bool {
	from, ok := destinationZone(g, playerID, fromZone)
	return ok && from.Remove(cardID)
}

func discardCardFromHand(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	return discardCardFromHandInBatch(g, playerID, cardID, 0)
}

func discardCardFromHandInBatch(g *game.Game, playerID game.PlayerID, cardID, simultaneousID id.ID) bool {
	player, ok := playerByID(g, playerID)
	if !ok || !player.Hand.Remove(cardID) {
		return false
	}
	card, cardOK := g.GetCardInstance(cardID)
	destination := zone.Graveyard
	shuffleIntoLibrary := false
	revealSource := false
	event := game.Event{}
	if cardOK {
		if _, ok := madnessCostForCard(cardFaceOrDefault(card, game.FaceFront)); ok {
			destination = zone.Exile
		}
		event = game.Event{
			Kind:           game.EventZoneChanged,
			Controller:     playerID,
			Player:         playerID,
			CardID:         cardID,
			FromZone:       zone.Hand,
			ToZone:         destination,
			SimultaneousID: simultaneousID,
		}
		replacement := replacementZoneChange(g, event)
		destination = replacement.destination
		destination = commanderReplacementDestination(g, card.ID, destination)
		shuffleIntoLibrary = replacement.shuffleIntoLibrary
		revealSource = replacement.revealSource
	}
	zoneOwner := playerID
	if destination == zone.Command && cardOK {
		zoneOwner = card.Owner
	}
	destinationCards, ok := destinationZone(g, zoneOwner, destination)
	if !ok {
		return false
	}
	revealZoneReplacementSource(g, event, revealSource)
	destinationCards.Add(cardID)
	shuffleLibraryIfRequested(g, destinationCards, destination, shuffleIntoLibrary)
	event = game.Event{
		Player:         playerID,
		CardID:         cardID,
		FromZone:       zone.Hand,
		ToZone:         destination,
		Amount:         1,
		SimultaneousID: simultaneousID,
	}
	event = emitZoneChangeEvent(g, event)
	// A command-zone replacement changes the destination, but the discard still happened.
	event.Kind = game.EventCardDiscarded
	emitEvent(g, event)
	return true
}

func shuffleLibraryIfRequested(g *game.Game, cards *zone.Zone, destination zone.Type, shuffle bool) {
	if shuffle && destination == zone.Library {
		cards.Shuffle(g.RNG)
	}
}

func emitPermanentLeaveEvents(g *game.Game, permanent *game.Permanent, controller game.PlayerID, destination zone.Type, simultaneousID id.ID) {
	event := game.Event{
		Controller:     controller,
		Player:         permanent.Owner,
		CardID:         permanent.CardInstanceID,
		Face:           permanent.Face,
		FaceDown:       permanent.FaceDown,
		PermanentID:    permanent.ObjectID,
		TokenName:      permanentTokenName(permanent),
		TokenDef:       permanent.TokenDef,
		FromZone:       zone.Battlefield,
		ToZone:         destination,
		SimultaneousID: simultaneousID,
	}
	if card, ok := g.GetCardInstance(event.CardID); ok {
		card.ZoneVersion++
		event.CardZoneVersion = card.ZoneVersion
	}
	event = emitZoneChangeEvent(g, event)
	if destination == zone.Graveyard {
		event.Kind = game.EventPermanentDied
		emitEvent(g, event)
	}
}

func destroyPermanent(g *game.Game, objectID id.ID) (*game.Permanent, bool) {
	return destroyPermanentInBatch(g, objectID, 0, false)
}

func destroyPermanentInBatch(g *game.Game, objectID, simultaneousID id.ID, preventRegeneration bool) (*game.Permanent, bool) {
	permanent, ok := permanentByObjectID(g, objectID)
	if !ok {
		return nil, false
	}
	if hasKeyword(g, permanent, game.Indestructible) {
		return nil, false
	}
	if replaceDestroyPermanent(g, permanent, preventRegeneration) {
		return nil, false
	}
	if commanderReplacementDestination(g, permanent.CardInstanceID, zone.Graveyard) == zone.Command {
		movePermanentToZoneInBatch(g, permanent, zone.Graveyard, simultaneousID)
		return nil, false
	}
	if !movePermanentToZoneInBatch(g, permanent, zone.Graveyard, simultaneousID) {
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
