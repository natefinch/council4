package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
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
	auraAttachment   opt.V[game.Target]
	auraEntryChoices map[game.ChoiceKey]game.ResolutionChoiceResult
	ForceTapped      bool
	KickerPaid       bool
	KickCount        int
	// Bargained makes the entering permanent record that its spell was bargained
	// as it was cast (CR 702.166b), feeding the "if it was bargained" enter
	// trigger's intervening-if. It is false for every permanent that did not
	// enter from a bargained cast.
	Bargained bool
	// OffspringPaid makes the entering permanent record that its spell's
	// Offspring additional cost was paid as it was cast (CR 702.171b), feeding
	// the linked "when this creature enters, if its offspring cost was paid" ETB
	// trigger's intervening-if. It is false for every permanent that did not
	// enter from a spell with that cost paid, including ordinary token copies
	// made by the trigger, so those tokens do not recurse. Spell copies inherit
	// the announced additional cost.
	OffspringPaid bool
	Evoked        bool
	// Dashed makes the entering permanent record that its spell was cast for its
	// Dash alternative cost (CR 702.109), feeding the Dash trigger's intervening
	// "if its dash cost was paid" condition. It is false for every permanent that
	// did not enter from a dashed cast.
	Dashed bool
	// Bestowed makes the entering permanent a bestowed Aura (CR 702.103b): it
	// records that the resolving spell was cast bestowed so the permanent's
	// Bestow static ability changes it from a creature into an Aura and it stays
	// attached to the creature it enchants. It is false for every permanent that
	// did not enter from a bestowed cast.
	Bestowed             bool
	WasCast              bool
	HasSpellAnnouncement bool
	CastController       game.PlayerID
	HasCastController    bool
	CastFromZone         zone.Type
	Counters             []game.CounterPlacement
	// SourceFaceDown survives the source zone's removal of face-down metadata.
	SourceFaceDown bool
	SimultaneousID id.ID
	XValue         int
	// EntersTransformed makes a transforming double-faced card enter the
	// battlefield converted (as its back face), backing the Transformers "More
	// Than Meets the Eye" alternative cast and "return it to the battlefield
	// converted" (CR 712). It is honored only for a transforming double-faced
	// card entering from its front face; every other permanent ignores it.
	EntersTransformed bool
	// ColorsOfManaSpentToCast carries the number of distinct colors of mana
	// spent to cast the spell that is resolving into this permanent, so a
	// Converge enters-with-counters replacement ("for each color of mana spent
	// to cast it") reads the count as the permanent enters. It is zero for a
	// permanent that did not enter from a cast spell (a token, a copy, a
	// put-into-play effect).
	ColorsOfManaSpentToCast int
	// ManaSpentByColorToCast carries, per color, how much colored mana was spent
	// to cast the spell that is resolving into this permanent, so an Adamant
	// enters-with-counters replacement ("if at least three <color> mana was spent
	// to cast this spell") reads it as the permanent enters. It is nil for a
	// permanent that did not enter from a cast spell (a token, a copy, a
	// put-into-play effect).
	ManaSpentByColorToCast map[color.Color]int
	// ManaSpentToCast carries the total amount of mana spent to cast the spell
	// that is resolving into this permanent, so Mockingbird's "with mana value
	// less than or equal to the amount of mana spent to cast this creature"
	// enters-as-copy filter reads it as the permanent enters. It is zero for a
	// permanent that did not enter from a cast spell (a token, a copy, a
	// put-into-play effect).
	ManaSpentToCast int
	// PutByAbilitySource records the object identity of the ability source that
	// put this permanent onto the battlefield from a choose-from-zone effect, so
	// the entering event can carry it for the "if it wasn't put onto the
	// battlefield with this ability" enter-trigger intervening-if (Kodama of the
	// East Tree). It is zero for every entry that carries no origin ability.
	PutByAbilitySource id.ID
}

func (options permanentCreationOptions) entryContext(e *Engine, agents [game.NumPlayers]PlayerAgent, log *TurnLog) enterBattlefieldContext {
	return enterBattlefieldContext{
		engine: e, agents: agents, log: log,
		xValue: options.XValue, hasSpellAnnouncement: options.WasCast || options.HasSpellAnnouncement,
		kickCount: options.KickCount, kickerPaid: options.KickerPaid, bargained: options.Bargained,
		wasCast: options.WasCast, castController: options.CastController, hasCastController: options.HasCastController,
		castFromZone: options.CastFromZone, colorsOfManaSpent: options.ColorsOfManaSpentToCast,
		manaSpentByColor: options.ManaSpentByColorToCast, manaSpentToCast: options.ManaSpentToCast,
	}
}

// createCardPermanentFaceWithOptions puts a card onto the battlefield as a new
// permanent (CR 400.7: it becomes a new object with no memory of its prior
// existence), assigning it a fresh object ID and timestamp. It applies the
// entering-the-battlefield replacement effects (CR 614.1c-d, e.g. "as this
// enters" choices and entering tapped or with counters) before the permanent
// exists, and the entering event then lets "when/whenever this enters" abilities
// trigger (CR 603.6a).
func createCardPermanentFaceWithOptions(e *Engine, g *game.Game, card *game.CardInstance, controller game.PlayerID, fromZone zone.Type, face game.FaceIndex, continuous []game.ContinuousEffect, options permanentCreationOptions, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (*game.Permanent, bool) {
	if departedPlayer(g, controller) || card != nil && departedPlayer(g, card.Owner) {
		return nil, false
	}
	enteringTransformed := false
	if options.EntersTransformed && face == game.FaceFront &&
		card != nil && card.Def != nil &&
		card.Def.IsTransformingDoubleFaced() && card.Def.Back.Exists {
		face = game.FaceBack
		enteringTransformed = true
	}
	faceDef, ok := cardFaceDef(card, face)
	if !ok {
		return nil, false
	}
	if entryFromZoneProhibited(g, faceDef, fromZone) {
		return nil, false
	}
	castXValue := 0
	if options.WasCast || options.HasSpellAnnouncement {
		castXValue = options.XValue
	}
	objectID := g.IDGen.Next()
	permanent := &game.Permanent{
		ObjectID:             objectID,
		CardInstanceID:       card.ID,
		Owner:                card.Owner,
		Controller:           controller,
		EnteredFromCast:      options.WasCast,
		CastXValue:           castXValue,
		HasSpellAnnouncement: options.WasCast || options.HasSpellAnnouncement,
		Face:                 face,
		Transformed:          enteringTransformed,
		SummoningSick:        true,
		Prepared:             faceDef.EntersPrepared,
	}
	initializePermanentCounters(permanent, faceDef)
	permanent.Bestowed = options.Bestowed
	applyInitialContinuousEffects(g, permanent, continuous)
	registerPermanentReplacementEffects(g, permanent)
	initializeReadAhead(e, g, permanent, agents, log)
	if options.ForceTapped {
		permanent.Tapped = true
	}
	destination := applyEnterBattlefieldReplacementEffects(options.entryContext(e, agents, log), g, permanent, fromZone)
	if destination != zone.Battlefield {
		redirectPendingCardEntry(g, card, permanent, fromZone, destination, options)
		return nil, false
	}
	for _, placement := range options.Counters {
		amount := replacementPermanentCounterPlacementAmount(g, controller, permanent, placement.Kind, placement.Amount)
		permanent.Counters.Add(placement.Kind, amount)
	}
	if !commitAuraEffectAttachment(g, permanent, options) {
		return nil, false
	}
	retained := retainedEntryCounters(g, card, fromZone, face, options.SourceFaceDown)
	addRetainedCounters(permanent, retained)
	g.Battlefield = append(g.Battlefield, permanent)
	if lore := permanent.Counters.Get(counter.Lore) - retained.Get(counter.Lore); lore > 0 {
		emitCounterAddedEvent(g, permanent, effectiveController(g, permanent), counter.Lore, retained.Get(counter.Lore), lore)
	}
	event := game.Event{
		SourceID:                  card.ID,
		Controller:                controller,
		Player:                    card.Owner,
		CardID:                    card.ID,
		Face:                      face,
		KickerPaid:                options.KickerPaid,
		Bargained:                 options.Bargained,
		OffspringPaid:             options.OffspringPaid,
		EnterEvoked:               options.Evoked,
		EnterDashed:               options.Dashed,
		EnterWasCast:              options.WasCast,
		EnterHasSpellAnnouncement: permanent.HasSpellAnnouncement,
		EnterCastController:       options.CastController,
		EnterHasCastController:    options.HasCastController,
		EnterCastFromZone:         options.CastFromZone,
		EnterXValue:               castXValue,
		PermanentID:               objectID,
		FromZone:                  fromZone,
		ToZone:                    zone.Battlefield,
		SimultaneousID:            options.SimultaneousID,
		EnterPutByAbilitySource:   options.PutByAbilitySource,
	}
	event = emitZoneChangeEvent(g, event)
	event.Kind = game.EventPermanentEnteredBattlefield
	emitEvent(g, event)
	return permanent, true
}

type preparedCardPermanentEntry struct {
	card        *game.CardInstance
	permanent   *game.Permanent
	controller  game.PlayerID
	fromZone    zone.Type
	options     permanentCreationOptions
	retained    counter.Set
	auraHost    *game.Permanent
	destination zone.Type
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
	if departedPlayer(g, controller) || card != nil && departedPlayer(g, card.Owner) {
		return preparedCardPermanentEntry{}, false
	}
	faceDef, ok := cardFaceDef(card, face)
	if !ok {
		return preparedCardPermanentEntry{}, false
	}
	if entryFromZoneProhibited(g, faceDef, fromZone) {
		return preparedCardPermanentEntry{}, false
	}
	castXValue := 0
	if options.WasCast || options.HasSpellAnnouncement {
		castXValue = options.XValue
	}
	permanent := &game.Permanent{
		ObjectID:             g.IDGen.Next(),
		CardInstanceID:       card.ID,
		Owner:                card.Owner,
		Controller:           controller,
		EnteredFromCast:      options.WasCast,
		CastXValue:           castXValue,
		HasSpellAnnouncement: options.WasCast || options.HasSpellAnnouncement,
		Face:                 face,
		SummoningSick:        true,
		Prepared:             faceDef.EntersPrepared,
	}
	initializePermanentCounters(permanent, faceDef)
	permanent.Bestowed = options.Bestowed
	// CR 611.2e characteristics participate in the entry replacement check.
	// The entrant itself remains outside the battlefield until the batch commits.
	applyInitialContinuousEffects(g, permanent, continuous)
	initializeReadAhead(e, g, permanent, agents, log)
	if options.ForceTapped {
		permanent.Tapped = true
	}
	destination := applyEnterBattlefieldReplacementEffects(options.entryContext(e, agents, log), g, permanent, fromZone)
	if destination != zone.Battlefield {
		discardPendingEntryContinuousEffects(g, permanent)
		return preparedCardPermanentEntry{card: card, permanent: permanent, destination: destination}, true
	}
	for _, placement := range options.Counters {
		amount := replacementPermanentCounterPlacementAmount(g, controller, permanent, placement.Kind, placement.Amount)
		permanent.Counters.Add(placement.Kind, amount)
	}
	if !prepareAuraEffectAttachment(g, permanent, options) {
		discardPendingEntryContinuousEffects(g, permanent)
		return preparedCardPermanentEntry{}, false
	}
	var auraHost *game.Permanent
	if permanent.AttachedTo.Exists {
		auraHost, _ = permanentByObjectID(g, permanent.AttachedTo.Val)
	}
	return preparedCardPermanentEntry{
		card:        card,
		permanent:   permanent,
		controller:  controller,
		fromZone:    fromZone,
		options:     options,
		retained:    retainedEntryCounters(g, card, fromZone, face, options.SourceFaceDown),
		auraHost:    auraHost,
		destination: destination,
	}, true
}

func commitSimultaneousCardPermanentEntries(g *game.Game, entries []preparedCardPermanentEntry) {
	for i := range entries {
		if host := entries[i].auraHost; host != nil {
			host.Attachments = append(host.Attachments, entries[i].permanent.ObjectID)
		}
	}
	for i := range entries {
		entry := &entries[i]
		registerPermanentReplacementEffects(g, entry.permanent)
	}
	for i := range entries {
		addRetainedCounters(entries[i].permanent, entries[i].retained)
		g.Battlefield = append(g.Battlefield, entries[i].permanent)
	}
	events := make([]game.Event, len(entries))
	for i := range entries {
		events[i] = simultaneousCardPermanentEntryEvent(&entries[i])
	}
	for i := range entries {
		entry := &entries[i]
		permanent := entry.permanent
		if lore := permanent.Counters.Get(counter.Lore) - entry.retained.Get(counter.Lore); lore > 0 {
			emitCounterAddedEvent(g, permanent, effectiveController(g, permanent), counter.Lore, entry.retained.Get(counter.Lore), lore)
		}
		event := emitZoneChangeEvent(g, events[i])
		event.Kind = game.EventPermanentEnteredBattlefield
		emitEnteredBattlefieldEvent(g, event, enteredBattlefieldEvents(events[i+1:]))
	}
}

func enteredBattlefieldEvents(events []game.Event) []game.Event {
	entered := make([]game.Event, len(events))
	for i, event := range events {
		event.Kind = game.EventPermanentEnteredBattlefield
		entered[i] = event
	}
	return entered
}

func simultaneousCardPermanentEntryEvent(entry *preparedCardPermanentEntry) game.Event {
	permanent := entry.permanent
	return game.Event{
		SourceID:                  entry.card.ID,
		Controller:                entry.controller,
		Player:                    entry.card.Owner,
		CardID:                    entry.card.ID,
		Face:                      permanent.Face,
		KickerPaid:                entry.options.KickerPaid,
		Bargained:                 entry.options.Bargained,
		OffspringPaid:             entry.options.OffspringPaid,
		EnterEvoked:               entry.options.Evoked,
		EnterDashed:               entry.options.Dashed,
		EnterWasCast:              entry.options.WasCast,
		EnterHasSpellAnnouncement: permanent.HasSpellAnnouncement,
		EnterCastController:       entry.options.CastController,
		EnterHasCastController:    entry.options.HasCastController,
		EnterCastFromZone:         entry.options.CastFromZone,
		EnterXValue:               permanent.CastXValue,
		PermanentID:               permanent.ObjectID,
		FromZone:                  entry.fromZone,
		ToZone:                    zone.Battlefield,
		SimultaneousID:            entry.options.SimultaneousID,
		EnterPutByAbilitySource:   entry.options.PutByAbilitySource,
	}
}

func discardPendingEntryContinuousEffects(g *game.Game, permanent *game.Permanent) {
	g.ContinuousEffects = slices.DeleteFunc(g.ContinuousEffects, func(effect game.ContinuousEffect) bool {
		return effect.AffectedObjectID == permanent.ObjectID
	})
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
	destination := applyEnterBattlefieldReplacementEffects(enterBattlefieldContext{
		engine:            e,
		agents:            agents,
		log:               log,
		wasCast:           wasCast,
		castController:    controller,
		hasCastController: wasCast,
		castFromZone:      fromZone,
	}, g, permanent, fromZone)
	if destination != zone.Battlefield {
		redirectPendingCardEntry(g, card, permanent, fromZone, destination, permanentCreationOptions{SourceFaceDown: true})
		return nil, false
	}
	addRetainedCounters(permanent, retainedEntryCounters(g, card, fromZone, face, fromZone == zone.Stack))
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
	retained          counter.Set
}

// preparePermanentZoneMove computes everything needed to move a permanent off the
// battlefield without yet moving any objects: it snapshots the permanent's last
// known information (so leaves-the-battlefield abilities and other look-back
// effects see its prior characteristics, CR 603.10, CR 608.2h), applies
// zone-change replacement effects to find the real destination (CR 614), and
// resolves merged-permanent component moves. The moved card becomes a new object
// in its new zone (CR 400.7).
func preparePermanentZoneMove(g *game.Game, permanent *game.Permanent, destination zone.Type) (preparedPermanentZoneMove, bool) {
	move, ok := preparePermanentZoneDeparture(g, permanent, destination)
	if !ok {
		return preparedPermanentZoneMove{}, false
	}
	moves := []preparedPermanentZoneMove{move}
	replacePermanentZoneMoveDestinations(g, moves)
	ok = finishPermanentZoneMove(g, &moves[0])
	return moves[0], ok
}

// preparePermanentZoneDeparture captures the pre-move state without asking any
// replacement questions. The whole batch's pending decisions are ordered APNAP.
func preparePermanentZoneDeparture(g *game.Game, permanent *game.Permanent, destination zone.Type) (preparedPermanentZoneMove, bool) {
	if _, ok := permanentByObjectID(g, permanent.ObjectID); !ok {
		return preparedPermanentZoneMove{}, false
	}
	snapshot := snapshotPermanent(g, permanent, zone.Battlefield)
	event := game.Event{
		Kind:               game.EventZoneChanged,
		Controller:         effectiveController(g, permanent),
		Player:             permanent.Owner,
		CardID:             permanent.CardInstanceID,
		Face:               permanent.Face,
		FaceDown:           permanent.FaceDown,
		PermanentID:        permanent.ObjectID,
		SubjectGoaded:      isGoadedNow(g, permanent),
		SubjectGoadedKnown: true,
		TokenName:          permanentTokenName(permanent),
		TokenDef:           permanent.TokenDef,
		FromZone:           zone.Battlefield,
		ToZone:             destination,
	}
	event.ControlledTriggerDoublers = captureControlledTriggerDoublerSnapshot(g)
	componentMoves, ok := mergedComponentZoneMoves(g, permanent, destination)
	if !ok {
		return preparedPermanentZoneMove{}, false
	}
	return preparedPermanentZoneMove{
		permanent:         permanent,
		snapshot:          snapshot,
		event:             event,
		actualDestination: destination,
		componentMoves:    componentMoves,
	}, true
}

func replacePermanentZoneMoveDestinations(g *game.Game, moves []preparedPermanentZoneMove) {
	type progress struct {
		ordinary      zoneChangeReplacement
		ordinaryDone  bool
		commanderDone bool
		componentDone []bool
	}
	type pendingChoice struct {
		player  game.PlayerID
		resolve func()
	}
	states := make([]progress, len(moves))
	for i := range moves {
		states[i].ordinary = newZoneChangeReplacement(moves[i].event)
		states[i].componentDone = make([]bool, len(moves[i].componentMoves))
	}
	for {
		var choices []pendingChoice
		addCommanderChoice := func(cardID id.ID, destination *zone.Type, done *bool) {
			if *done || !isCommanderCardID(g, cardID) || (*destination != zone.Hand && *destination != zone.Library) {
				return
			}
			card, ok := g.GetCardInstance(cardID)
			if !ok || !isPlayerAlive(g, card.Owner) {
				return
			}
			choices = append(choices, pendingChoice{card.Owner, func() {
				*destination = commanderReplacementDestination(g, cardID, *destination)
				*done = true
			}})
		}
		for i := range moves {
			move := &moves[i]
			state := &states[i]
			if !state.ordinaryDone {
				matches := state.ordinary.pendingChoices(g)
				if len(matches) != 0 {
					choices = append(choices, pendingChoice{replacementEventPlayer(move.event), func() {
						decision := chooseReplacementDecision(g, replacementEventPlayer(move.event), replacementEffectLabels(matches))
						state.ordinary.apply(selectedReplacementEffect(matches, decision))
					}})
					continue
				}
				state.ordinaryDone = true
				move.replacement = state.ordinary.result
				move.actualDestination = move.replacement.destination
				for j := range move.componentMoves {
					move.componentMoves[j].destination = move.replacement.destination
				}
			}
			// Ordinary replacements use the controller (CR 616.1); the
			// command-zone option remains the owner's decision (CR 903.9b).
			if !move.permanent.Token {
				addCommanderChoice(move.permanent.CardInstanceID, &move.actualDestination, &state.commanderDone)
			}
			for j := range move.componentMoves {
				component := &move.componentMoves[j]
				if component.tokenDef == nil {
					addCommanderChoice(component.cardID, &component.destination, &state.componentDone[j])
				}
			}
		}
		if len(choices) == 0 {
			return
		}
		next := choices[0]
	findNext:
		for _, player := range triggerAPNAPPlayers(g) {
			for _, choice := range choices {
				if choice.player == player {
					next = choice
					break findNext
				}
			}
		}
		next.resolve()
		// A replacement may expose a new decision for an earlier player.
		// Reconsider every outstanding choice from the active player (101.4d).
	}
}

func finishPermanentZoneMove(g *game.Game, move *preparedPermanentZoneMove) bool {
	permanent := move.permanent
	if _, ok := destinationZone(g, permanent.Owner, move.actualDestination); !ok {
		return false
	}
	if !permanent.Counters.IsEmpty() && retainsCounters(g, permanentEffectiveAbilitiesView(g, permanent),
		conditionContext{controller: move.event.Controller, source: permanent}, zone.Battlefield, move.actualDestination) {
		move.retained = cloneCounters(permanent.Counters)
	}
	for i := range move.componentMoves {
		component := &move.componentMoves[i]
		if _, ok := destinationZone(g, component.owner, component.destination); !ok {
			return false
		}
		if !permanent.Counters.IsEmpty() && retainsCounters(g, permanentEffectiveAbilitiesView(g, permanent),
			conditionContext{controller: move.event.Controller, source: permanent}, zone.Battlefield, component.destination) {
			component.retained = cloneCounters(permanent.Counters)
		}
	}
	return true
}

// applyPreparedPermanentZoneMove commits a prepared permanent zone move: it
// records last known information, detaches the permanent and its attachments,
// removes it from the battlefield, and places the underlying card (or token) into
// the destination zone; a card going to a library, graveyard, or hand goes to its
// owner's (CR 400.3). The permanent ceases to exist and the card becomes a new
// object in its new zone (CR 400.7); a token that leaves the battlefield ceases to
// exist as a state-based action shortly after (CR 111.7).
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
		storeDepartingCounters(g, removed.ObjectID, move.retained)
		destinationCards.Add(removed.ObjectID)
		move.event = emitPermanentLeaveEvents(g, removed, move.event.Controller, move.event.SubjectGoaded, move.event.SubjectGoadedKnown, move.actualDestination, move.event.SimultaneousID, move.event.ControlledTriggerDoublers)
	} else {
		storeDepartingCounters(g, removed.CardInstanceID, move.retained)
		destinationCards.Add(removed.CardInstanceID)
		shuffleLibraryIfRequested(g, destinationCards, move.actualDestination, move.replacement.shuffleIntoLibrary)
		placeRedirectExileCounter(g, removed.Owner, removed.CardInstanceID, move.replacement)
		move.event = emitPermanentLeaveEvents(g, removed, move.event.Controller, move.event.SubjectGoaded, move.event.SubjectGoadedKnown, move.actualDestination, move.event.SimultaneousID, move.event.ControlledTriggerDoublers)
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
		storeDepartingCounters(g, component.cardID, component.retained)
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
	return movePermanentToZoneInBatchWithControlledTriggerSnapshot(g, permanent, destination, simultaneousID, nil)
}

func movePermanentToZoneInBatchWithControlledTriggerSnapshot(
	g *game.Game,
	permanent *game.Permanent,
	destination zone.Type,
	simultaneousID id.ID,
	triggerSnapshot *game.ControlledTriggerDoublerSnapshot,
) bool {
	move, ok := preparePermanentZoneMove(g, permanent, destination)
	if !ok {
		return false
	}
	move.event.SimultaneousID = simultaneousID
	if triggerSnapshot != nil {
		move.event.ControlledTriggerDoublers = triggerSnapshot
	}
	moved := applyPreparedPermanentZoneMove(g, &move)
	settleDepartedControl(g)
	return moved
}

// movePermanentsToZoneSimultaneously moves several permanents to the same zone as
// a single simultaneous event so look-back-in-time abilities see them all leaving
// together (CR 603.10a) and so a single damage/zone-change batch is produced. All
// moves are prepared (last known information snapshotted) before any is applied so
// each one's replacement and trigger checks use the pre-move game state.
// Replacement choices are staged together in APNAP order, without changing the
// input order of committed moves or their correlated results. CR 404.3
// lets the owner arrange cards put into the same graveyard at once; this engine
// does not prompt for that order and adds them in processing order instead.
func movePermanentsToZoneSimultaneously(g *game.Game, permanents []*game.Permanent, destination zone.Type) bool {
	return movePermanentsToZoneSimultaneouslyInBatch(g, permanents, destination, 0)
}

func movePermanentsToZoneSimultaneouslyInBatch(g *game.Game, permanents []*game.Permanent, destination zone.Type, simultaneousID id.ID) bool {
	results := movePermanentsToZoneSimultaneouslyWithResultsInBatch(g, permanents, destination, simultaneousID)
	for _, result := range results {
		if result.moved {
			return true
		}
	}
	return false
}

type permanentZoneMoveResult struct {
	permanent   *game.Permanent
	destination zone.Type
	moved       bool
	arrivals    []game.ZoneCardSnapshot
}

// movePermanentsToZoneSimultaneouslyWithResults is the result-bearing form of
// movePermanentsToZoneSimultaneously. It preserves one result per prepared move
// so linked effects can distinguish a requested destination from a replacement
// destination while retaining one simultaneous batch.
func movePermanentsToZoneSimultaneouslyWithResults(g *game.Game, permanents []*game.Permanent, destination zone.Type) []permanentZoneMoveResult {
	return movePermanentsToZoneSimultaneouslyWithResultsInBatch(g, permanents, destination, 0)
}

func movePermanentsToZoneSimultaneouslyWithResultsInBatch(g *game.Game, permanents []*game.Permanent, destination zone.Type, simultaneousID id.ID) []permanentZoneMoveResult {
	moves := preparePermanentZoneMovesInBatch(g, permanents, destination, simultaneousID)
	return applyPreparedPermanentZoneMoves(g, moves)
}

func preparePermanentZoneMovesInBatch(g *game.Game, permanents []*game.Permanent, destination zone.Type, simultaneousID id.ID) []preparedPermanentZoneMove {
	moves := make([]preparedPermanentZoneMove, 0, len(permanents))
	for _, permanent := range permanents {
		move, ok := preparePermanentZoneDeparture(g, permanent, destination)
		if ok {
			moves = append(moves, move)
		}
	}
	replacePermanentZoneMoveDestinations(g, moves)
	prepared := moves[:0]
	for i := range moves {
		if finishPermanentZoneMove(g, &moves[i]) {
			prepared = append(prepared, moves[i])
		}
	}
	moves = prepared
	if len(moves) > 0 {
		if simultaneousID == 0 {
			simultaneousID = g.IDGen.Next()
		}
		for i := range moves {
			moves[i].event.SimultaneousID = simultaneousID
		}
	}
	return moves
}

func applyPreparedPermanentZoneMoves(g *game.Game, moves []preparedPermanentZoneMove) []permanentZoneMoveResult {
	results := make([]permanentZoneMoveResult, 0, len(moves))
	for i := range moves {
		result := permanentZoneMoveResult{
			permanent:   moves[i].permanent,
			destination: moves[i].actualDestination,
			moved:       applyPreparedPermanentZoneMove(g, &moves[i]),
		}
		if result.moved {
			for _, component := range moves[i].snapshot.ZoneCards {
				if card, ok := g.GetCardInstance(component.CardID); ok {
					result.arrivals = append(result.arrivals, game.ZoneCardSnapshot{
						CardID: card.ID, ZoneVersion: card.ZoneVersion,
					})
				}
			}
		}
		results = append(results, result)
	}
	settleDepartedControl(g)
	return results
}

type mergedCardZoneMove struct {
	cardID       id.ID
	face         game.FaceIndex
	faceDown     bool
	faceDownFace game.FaceIndex
	owner        game.PlayerID
	destination  zone.Type
	tokenDef     *game.CardDef
	retained     counter.Set
}

func mergedComponentZoneMoves(g *game.Game, permanent *game.Permanent, destination zone.Type) ([]mergedCardZoneMove, bool) {
	moves := make([]mergedCardZoneMove, 0, len(permanent.MergedCards))
	for _, component := range permanent.MergedCards {
		faceDown := component.FaceDown || permanent.FaceDown
		faceDownFace := component.FaceDownFace
		if permanent.FaceDown && !component.FaceDown {
			faceDownFace = component.Face
		}
		if component.TokenDef != nil {
			moves = append(moves, mergedCardZoneMove{
				face:         component.Face,
				faceDown:     faceDown,
				faceDownFace: faceDownFace,
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
		moves = append(moves, mergedCardZoneMove{
			cardID:       card.ID,
			face:         component.Face,
			faceDown:     faceDown,
			faceDownFace: faceDownFace,
			owner:        card.Owner,
			destination:  destination,
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

type zoneMoveOptions struct {
	bottom        bool
	faceDownExile bool
}

// moveCardBetweenZonesInBatch moves a card from one zone to another, applying any
// zone-change replacement effects to determine its real destination (CR 614) and
// commander command-zone handling (the hand/library-to-command move is a
// replacement effect, CR 903.9b; the graveyard/exile case is a state-based action,
// CR 903.9a). The card becomes a new object in its new zone (CR 400.7).
func moveCardBetweenZonesInBatch(g *game.Game, playerID game.PlayerID, cardID id.ID, fromZone, toZone zone.Type, bottom bool, simultaneousID id.ID) bool {
	_, moved := moveCardBetweenZonesWithEvent(g, game.Event{
		Controller: playerID, Player: playerID, CardID: cardID,
		FromZone: fromZone, ToZone: toZone, SimultaneousID: simultaneousID,
	}, zoneMoveOptions{bottom: bottom})
	return moved
}

// The source remains intact while replacements and the owner's choices run.
// Graveyard/exile arrivals are recorded here, not processed as immediate SBAs.
func moveCardBetweenZonesWithEvent(g *game.Game, event game.Event, options zoneMoveOptions) (zone.Type, bool) {
	card, ok := g.GetCardInstance(event.CardID)
	from, valid := destinationZone(g, event.Player, event.FromZone)
	if !ok || !valid || !from.Contains(event.CardID) {
		return zone.None, false
	}
	version := card.ZoneVersion
	event.Kind = game.EventZoneChanged
	replacement := replacementZoneChange(g, event)
	replacement.destination = commanderReplacementDestination(g, card.ID, replacement.destination)
	if card.ZoneVersion != version || !from.Contains(card.ID) {
		return zone.None, false
	}
	moved := moveCardBetweenZonesAfterReplacement(g, event.Player, card.ID, event.FromZone,
		replacement, event, options, event.SimultaneousID)
	return replacement.destination, moved
}

func moveCardBetweenZonesAfterReplacement(
	g *game.Game,
	playerID game.PlayerID,
	cardID id.ID,
	fromZone zone.Type,
	replacement zoneChangeReplacementResult,
	event game.Event,
	options zoneMoveOptions,
	simultaneousID id.ID,
) bool {
	destination := replacement.destination
	from, ok := destinationZone(g, playerID, fromZone)
	if !ok || !from.Contains(cardID) {
		return false
	}
	if destination == fromZone && destination != zone.Exile {
		revealZoneReplacementSource(g, event, replacement.revealSource)
		shuffleLibraryIfRequested(g, from, destination, replacement.shuffleIntoLibrary)
		return true
	}
	sourceFaceDown := event.FaceDown || from.IsFaceDown(cardID)
	if !from.Remove(cardID) {
		return false
	}
	if !placeMovedCardAfterReplacement(g, playerID, cardID, fromZone, replacement, event, sourceFaceDown, options, simultaneousID) {
		from.Add(cardID)
		from.SetFaceDown(cardID, sourceFaceDown)
		return false
	}
	return true
}

// Entry callers may already have removed the resolving or moving card from its
// source zone. Complete that pending move without publishing it back to a zone.
func redirectPendingCardEntry(g *game.Game, card *game.CardInstance, permanent *game.Permanent, fromZone, destination zone.Type, options permanentCreationOptions) {
	event := game.Event{
		Kind: game.EventZoneChanged, Controller: effectiveController(g, permanent), Player: card.Owner,
		CardID: card.ID, Face: permanent.Face, FaceDown: options.SourceFaceDown,
		FromZone: fromZone, ToZone: destination, SimultaneousID: options.SimultaneousID,
	}
	replacement := replacementZoneChange(g, event)
	replacement.destination = commanderReplacementDestination(g, card.ID, replacement.destination)
	if from, ok := destinationZone(g, card.Owner, fromZone); ok && from.Contains(card.ID) {
		if !moveCardBetweenZonesAfterReplacement(g, card.Owner, card.ID, fromZone, replacement, event, zoneMoveOptions{}, options.SimultaneousID) {
			panic(UnsupportedError{Reason: "cannot complete replaced battlefield entry"})
		}
		return
	}
	if !placeMovedCardAfterReplacement(g, card.Owner, card.ID, fromZone, replacement, event, options.SourceFaceDown, zoneMoveOptions{}, options.SimultaneousID) {
		panic(UnsupportedError{Reason: "cannot complete replaced battlefield entry"})
	}
}

func placeMovedCardAfterReplacement(
	g *game.Game,
	playerID game.PlayerID,
	cardID id.ID,
	fromZone zone.Type,
	replacement zoneChangeReplacementResult,
	event game.Event,
	sourceFaceDown bool,
	options zoneMoveOptions,
	simultaneousID id.ID,
) bool {
	destination := replacement.destination
	zoneOwner := playerID
	if destination == zone.Command {
		if card, ok := g.GetCardInstance(cardID); ok {
			zoneOwner = card.Owner
		}
	}
	to, ok := destinationZone(g, zoneOwner, destination)
	if !ok {
		return false
	}
	revealZoneReplacementSource(g, event, replacement.revealSource)
	if options.bottom && destination == zone.Library {
		to.AddToBottom(cardID)
	} else {
		to.Add(cardID)
	}
	if options.faceDownExile && destination == zone.Exile {
		to.SetFaceDown(cardID, true)
	}
	shuffleLibraryIfRequested(g, to, destination, replacement.shuffleIntoLibrary)
	event.Player = playerID
	event.CardID = cardID
	event.FaceDown = sourceFaceDown || options.faceDownExile && destination == zone.Exile
	event.FromZone = fromZone
	event.ToZone = destination
	event.SimultaneousID = simultaneousID
	emitZoneChangeEvent(g, event)
	placeRedirectExileCounter(g, zoneOwner, cardID, replacement)
	return true
}

func removeCardFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, fromZone zone.Type) bool {
	from, ok := destinationZone(g, playerID, fromZone)
	return ok && from.Remove(cardID)
}

func discardCardFromHand(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	return discardCardFromHandInBatch(g, playerID, cardID, 0)
}

// discardCardsAtRandomFromHand discards amount cards chosen uniformly at random
// from the player's hand as one simultaneous batch (CR 701.9a). It discards
// min(amount, hand size) cards and returns the discarded cards' instance IDs in
// the order they left the hand. Randomness is drawn from g.RNG, so a replay with
// the same seed discards the same cards.
func discardCardsAtRandomFromHand(g *game.Game, playerID game.PlayerID, amount int) []id.ID {
	player, ok := playerByID(g, playerID)
	if !ok {
		return nil
	}
	candidates := player.Hand.All()
	amount = min(amount, len(candidates))
	if amount <= 0 {
		return nil
	}
	order := make([]int, len(candidates))
	for i := range order {
		order[i] = i
	}
	g.RNG.Shuffle(len(order), func(i, j int) {
		order[i], order[j] = order[j], order[i]
	})
	simultaneousID := g.IDGen.Next()
	var discarded []id.ID
	for _, idx := range order[:amount] {
		if discardCardFromHandInBatch(g, playerID, candidates[idx], simultaneousID) {
			discarded = append(discarded, candidates[idx])
		}
	}
	return discarded
}

// discardCardFromHandInBatch discards a card by moving it from its owner's hand to
// their graveyard (CR 701.9a), subject to zone-change replacement effects (e.g.
// madness exiles it instead, CR 702.35a). The discarded card becomes a new object
// in its destination zone (CR 400.7).
func discardCardFromHandInBatch(g *game.Game, playerID game.PlayerID, cardID, simultaneousID id.ID) bool {
	player, ok := playerByID(g, playerID)
	if !ok || !player.Hand.Remove(cardID) {
		return false
	}
	card, cardOK := g.GetCardInstance(cardID)
	destination := zone.Graveyard
	shuffleIntoLibrary := false
	revealSource := false
	var replacement zoneChangeReplacementResult
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
		replacement = replacementZoneChange(g, event)
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
	placeRedirectExileCounter(g, zoneOwner, cardID, replacement)
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

func emitPermanentLeaveEvents(g *game.Game, permanent *game.Permanent, controller game.PlayerID, subjectGoaded, subjectGoadedKnown bool, destination zone.Type, simultaneousID id.ID, doublers *game.ControlledTriggerDoublerSnapshot) game.Event {
	event := game.Event{
		Controller:                controller,
		Player:                    permanent.Owner,
		CardID:                    permanent.CardInstanceID,
		Face:                      permanent.Face,
		FaceDown:                  permanent.FaceDown,
		PermanentID:               permanent.ObjectID,
		SubjectGoaded:             subjectGoaded,
		SubjectGoadedKnown:        subjectGoadedKnown,
		TokenName:                 permanentTokenName(permanent),
		TokenDef:                  permanent.TokenDef,
		FromZone:                  zone.Battlefield,
		ToZone:                    destination,
		SimultaneousID:            simultaneousID,
		ControlledTriggerDoublers: doublers,
	}
	if card, ok := g.GetCardInstance(event.CardID); ok {
		card.ZoneVersion++
		event.CardZoneVersion = card.ZoneVersion
	}
	event = emitZoneChangeEvent(g, event)
	if destination == zone.Graveyard {
		death := event
		death.Kind = game.EventPermanentDied
		emitEvent(g, death)
	}
	return event
}

func destroyPermanent(g *game.Game, objectID id.ID) (*game.Permanent, bool) {
	return destroyPermanentInBatch(g, objectID, 0, false)
}

// destroyPermanentInBatch destroys a permanent by moving it from the battlefield
// to its owner's graveyard (CR 701.8a), unless it has indestructible (CR 702.12b)
// or the destruction is replaced by a shield counter (CR 122.1c) or regeneration
// (CR 614.8). A commander moved to the graveyard may be put into the command zone
// by its owner as a state-based action (CR 903.9a). Returns the destroyed
// permanent and whether it was actually destroyed.
func destroyPermanentInBatch(g *game.Game, objectID, simultaneousID id.ID, preventRegeneration bool) (*game.Permanent, bool) {
	permanent, destroyed, _ := destroyPermanentInBatchResult(g, objectID, simultaneousID, preventRegeneration)
	return permanent, destroyed
}

// destroyPermanentInBatchResult also reports whether destruction or its
// replacement changed game state, allowing state-based actions to repeat after a
// replacement removes an Aura or otherwise changes the battlefield.
func destroyPermanentInBatchResult(g *game.Game, objectID, simultaneousID id.ID, preventRegeneration bool) (permanent *game.Permanent, destroyed, changed bool) {
	return destroyPermanentInBatchResultWithBatch(g, objectID, simultaneousID, preventRegeneration, nil)
}

func destroyPermanentInBatchResultWithBatch(g *game.Game, objectID, simultaneousID id.ID, preventRegeneration bool, batch *destroyBatch) (permanent *game.Permanent, destroyed, changed bool) {
	permanent, ok := permanentByObjectID(g, objectID)
	if !ok {
		return nil, false, false
	}
	if hasKeyword(g, permanent, game.Indestructible) {
		return nil, false, false
	}
	if batch == nil {
		if replaced, changed := replaceDestroyPermanent(g, permanent, preventRegeneration, simultaneousID); replaced {
			return nil, false, changed
		}
	} else {
		replacement, replaced := planDestroyPermanentReplacement(g, permanent, preventRegeneration, simultaneousID, destructionFromEffect)
		if replaced {
			return nil, false, applyDestroyReplacement(g, permanent, replacement, batch)
		}
	}
	if !movePermanentToZoneInBatch(g, permanent, zone.Graveyard, simultaneousID) {
		return nil, false, false
	}
	return permanent, true, true
}

// destinationZone returns the zone object for a given owner and zone type.
// Library, hand, and graveyard belong to a specific player, so an object that
// would go to one of them goes to its owner's corresponding zone (CR 400.3). The
// exile and command zones are shared zones in the rules (CR 400.1); this engine
// represents them per owner, so callers pass the object's owner for those too.
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
