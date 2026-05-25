package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

type pendingTriggeredAbility struct {
	controller   game.PlayerID
	sourceID     id.ID
	sourceCardID id.ID
	sourceToken  *game.CardDef
	abilityIndex int
	targets      []game.Target
	event        game.GameEvent
}

func (e *Engine) putTriggeredAbilitiesOnStack(g *game.Game) bool {
	return e.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, nil)
}

func (e *Engine) putTriggeredAbilitiesOnStackWithChoices(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if g == nil {
		return false
	}
	start := g.TriggerEventCursor
	if start < 0 || start > len(g.Events) {
		start = len(g.Events)
	}
	events := append([]game.GameEvent(nil), g.Events[start:]...)
	g.TriggerEventCursor = len(g.Events)
	if len(events) == 0 {
		return false
	}

	pending := e.detectTriggeredAbilities(g, events)
	if len(pending) == 0 {
		return false
	}
	for _, trigger := range e.orderTriggeredAbilitiesAPNAP(g, pending, agents, log) {
		g.Stack.Push(&game.StackObject{
			ID:              g.IDGen.Next(),
			Kind:            game.StackTriggeredAbility,
			SourceID:        trigger.sourceID,
			SourceCardID:    trigger.sourceCardID,
			SourceTokenDef:  trigger.sourceToken,
			AbilityIndex:    trigger.abilityIndex,
			TriggerEvent:    trigger.event,
			HasTriggerEvent: true,
			Controller:      trigger.controller,
			Targets:         append([]game.Target(nil), trigger.targets...),
		})
	}
	return true
}

func (e *Engine) detectTriggeredAbilities(g *game.Game, events []game.GameEvent) []pendingTriggeredAbility {
	var pending []pendingTriggeredAbility
	for _, event := range events {
		for _, permanent := range g.Battlefield {
			pending = append(pending, e.detectTriggeredAbilitiesFromPermanent(g, permanent, event)...)
		}
		if source := leftBattlefieldTriggerSource(g, event); source != nil {
			pending = append(pending, e.detectTriggeredAbilitiesFromPermanent(g, source, event)...)
		}
	}
	return pending
}

func (e *Engine) detectTriggeredAbilitiesFromPermanent(g *game.Game, permanent *game.Permanent, event game.GameEvent) []pendingTriggeredAbility {
	if permanent == nil {
		return nil
	}
	def := permanentCardDef(g, permanent)
	if def == nil {
		return nil
	}
	var pending []pendingTriggeredAbility
	controller := effectiveController(g, permanent)
	for i := range def.Abilities {
		ability := &def.Abilities[i]
		if ability.Kind != game.TriggeredAbility || ability.Trigger == nil {
			continue
		}
		if !triggerMatchesEvent(g, permanent, ability.Trigger.Pattern, event) || !triggerInterveningIf(g, controller, ability.Trigger, &event) {
			continue
		}
		pending = append(pending, pendingTriggeredAbility{
			controller:   controller,
			sourceID:     permanent.ObjectID,
			sourceCardID: permanent.CardInstanceID,
			sourceToken:  permanent.TokenDef,
			abilityIndex: i,
			event:        event,
		})
	}
	return pending
}

func leftBattlefieldTriggerSource(g *game.Game, event game.GameEvent) *game.Permanent {
	if event.FromZone != game.ZoneBattlefield || event.PermanentID == 0 {
		return nil
	}
	if permanentByObjectID(g, event.PermanentID) != nil {
		return nil
	}
	if event.CardID != 0 {
		if card := g.GetCardInstance(event.CardID); card == nil || card.Def == nil {
			return nil
		}
		return &game.Permanent{
			ObjectID:       event.PermanentID,
			CardInstanceID: event.CardID,
			Owner:          event.Player,
			Controller:     event.Controller,
		}
	}
	if event.TokenDef == nil {
		return nil
	}
	return &game.Permanent{
		ObjectID:   event.PermanentID,
		Owner:      event.Player,
		Controller: event.Controller,
		Token:      true,
		TokenDef:   event.TokenDef,
	}
}

func (e *Engine) triggerTargets(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, ability *game.AbilityDef, agents [game.NumPlayers]PlayerAgent, log *TurnLog) ([]game.Target, bool) {
	choices := targetChoicesForAbilityFromSourceObject(g, controller, source, sourceObjectID, ability)
	if len(choices) == 0 {
		return nil, false
	}
	if len(choices) == 1 {
		return append([]game.Target(nil), choices[0]...), true
	}
	selected := e.chooseChoice(g, agents, targetChoiceRequest(controller, "Choose triggered ability targets.", choices), log)
	if len(selected) != 1 || selected[0] < 0 || selected[0] >= len(choices) {
		return append([]game.Target(nil), choices[0]...), true
	}
	return append([]game.Target(nil), choices[selected[0]]...), true
}

func triggerMatchesEvent(g *game.Game, source *game.Permanent, pattern game.TriggerPattern, event game.GameEvent) bool {
	if source == nil || pattern.Event == game.EventUnknown || pattern.Event != event.Kind {
		return false
	}

	// Trigger patterns are checked when the triggering event is processed, and
	// LTB/dies checks may need last-known information for the moved permanent
	// (CR 603.2, CR 603.6c, CR 603.10).
	sourceController := effectiveController(g, source)
	if !triggerControllerMatches(sourceController, pattern.Controller, event.Controller) {
		return false
	}
	if !triggerSourceMatches(source, pattern.Source, event) {
		return false
	}
	if !triggerPlayerMatches(sourceController, pattern.Player, event.Player) {
		return false
	}
	if pattern.MatchFromZone && pattern.FromZone != event.FromZone {
		return false
	}
	if pattern.MatchToZone && pattern.ToZone != event.ToZone {
		return false
	}
	if pattern.DamageRecipient != game.DamageRecipientNone && pattern.DamageRecipient != event.DamageRecipient {
		return false
	}
	if pattern.Event == game.EventBeginningOfStep && pattern.Step != game.StepNone && pattern.Step != event.Step {
		return false
	}
	if pattern.MatchPermanentType && !eventPermanentHasType(g, event, pattern.PermanentType) {
		return false
	}
	if !eventPermanentTypeFiltersMatch(g, event, pattern.RequirePermanentTypes, pattern.ExcludePermanentTypes) {
		return false
	}
	if !eventCardTypeFiltersMatch(g, event, pattern.RequireCardTypes, pattern.ExcludeCardTypes) {
		return false
	}
	return true
}

func triggerInterveningIf(g *game.Game, controller game.PlayerID, trigger *game.TriggerCondition, event *game.GameEvent) bool {
	if trigger == nil {
		return true
	}
	// Intervening "if" conditions are checked both as the event triggers and as
	// the ability resolves (CR 603.4).
	if trigger.InterveningIfControllerLifeAtLeast != 0 {
		player := playerByID(g, controller)
		if player == nil || player.Life < trigger.InterveningIfControllerLifeAtLeast {
			return false
		}
	}
	if trigger.InterveningIfEventPermanentHadCounters && !eventPermanentHadCounters(g, event) {
		return false
	}
	return true
}

func triggerControllerMatches(sourceController game.PlayerID, filter game.TriggerControllerFilter, eventController game.PlayerID) bool {
	switch filter {
	case game.TriggerControllerYou:
		return eventController == sourceController
	case game.TriggerControllerOpponent:
		return eventController != sourceController && eventController >= 0 && eventController < game.NumPlayers
	default:
		return true
	}
}

func triggerSourceMatches(source *game.Permanent, filter game.TriggerSourceFilter, event game.GameEvent) bool {
	if filter != game.TriggerSourceSelf {
		return true
	}
	return (source.ObjectID != 0 && event.SourceObjectID == source.ObjectID) ||
		(source.ObjectID != 0 && event.PermanentID == source.ObjectID) ||
		(source.CardInstanceID != 0 && event.SourceID == source.CardInstanceID) ||
		(source.CardInstanceID != 0 && event.CardID == source.CardInstanceID)
}

func triggerPlayerMatches(sourceController game.PlayerID, filter game.TriggerPlayerFilter, eventPlayer game.PlayerID) bool {
	switch filter {
	case game.TriggerPlayerYou:
		return eventPlayer == sourceController
	case game.TriggerPlayerOpponent:
		return eventPlayer != sourceController && eventPlayer >= 0 && eventPlayer < game.NumPlayers
	default:
		return true
	}
}

func eventPermanentHasType(g *game.Game, event game.GameEvent, cardType game.CardType) bool {
	if event.PermanentID != 0 {
		if permanent := permanentByObjectID(g, event.PermanentID); permanent != nil {
			return permanentHasType(g, permanent, cardType)
		}
		// Leaves-the-battlefield and dies triggers look back at the permanent's
		// last existence on the battlefield (CR 603.10).
		if snapshot, ok := lastKnownObject(g, event.PermanentID); ok {
			return slices.Contains(snapshot.Types, cardType)
		}
	}
	if event.CardID != 0 {
		if card := g.GetCardInstance(event.CardID); card != nil && card.Def != nil {
			return card.Def.HasType(cardType)
		}
	}
	if event.TokenDef != nil {
		return event.TokenDef.HasType(cardType)
	}
	return false
}

func eventPermanentTypeFiltersMatch(g *game.Game, event game.GameEvent, required []game.CardType, excluded []game.CardType) bool {
	for _, cardType := range required {
		if !eventPermanentHasType(g, event, cardType) {
			return false
		}
	}
	for _, cardType := range excluded {
		if eventPermanentHasType(g, event, cardType) {
			return false
		}
	}
	return true
}

func eventCardTypeFiltersMatch(g *game.Game, event game.GameEvent, required []game.CardType, excluded []game.CardType) bool {
	types := event.CardTypes
	if len(types) == 0 && event.CardID != 0 {
		if card := g.GetCardInstance(event.CardID); card != nil && card.Def != nil {
			types = card.Def.Types
		}
	}
	for _, cardType := range required {
		if !slices.Contains(types, cardType) {
			return false
		}
	}
	for _, cardType := range excluded {
		if slices.Contains(types, cardType) {
			return false
		}
	}
	return true
}

func eventPermanentHadCounters(g *game.Game, event *game.GameEvent) bool {
	if event == nil || event.PermanentID == 0 {
		return false
	}
	if permanent := permanentByObjectID(g, event.PermanentID); permanent != nil {
		return !permanent.Counters.IsEmpty()
	}
	if snapshot, ok := lastKnownObject(g, event.PermanentID); ok {
		return !snapshot.Counters.IsEmpty()
	}
	return false
}

func (e *Engine) orderTriggeredAbilitiesAPNAP(g *game.Game, triggers []pendingTriggeredAbility, agents [game.NumPlayers]PlayerAgent, log *TurnLog) []pendingTriggeredAbility {
	if len(triggers) == 0 || g == nil {
		return triggers
	}
	ordered := make([]pendingTriggeredAbility, 0, len(triggers))
	used := make([]bool, len(triggers))
	for _, playerID := range triggerAPNAPPlayers(g) {
		var playerTriggers []pendingTriggeredAbility
		for i, trigger := range triggers {
			if trigger.controller == playerID {
				playerTriggers = append(playerTriggers, trigger)
				used[i] = true
			}
		}
		ordered = append(ordered, e.preparePlayerTriggers(g, playerID, playerTriggers, agents, log)...)
	}
	for i, trigger := range triggers {
		if !used[i] {
			ordered = append(ordered, trigger)
		}
	}
	return ordered
}

func (e *Engine) preparePlayerTriggers(g *game.Game, playerID game.PlayerID, triggers []pendingTriggeredAbility, agents [game.NumPlayers]PlayerAgent, log *TurnLog) []pendingTriggeredAbility {
	ordered := e.chooseTriggerOrder(g, playerID, triggers, agents, log)
	prepared := make([]pendingTriggeredAbility, 0, len(ordered))
	for _, trigger := range ordered {
		source := pendingTriggerSourceDef(g, trigger)
		targets, ok := e.triggerTargets(g, trigger.controller, source, trigger.sourceID, pendingTriggerAbilityFromDef(source, trigger), agents, log)
		if !ok {
			continue
		}
		trigger.targets = targets
		prepared = append(prepared, trigger)
	}
	return prepared
}

func pendingTriggerAbility(g *game.Game, trigger pendingTriggeredAbility) *game.AbilityDef {
	return pendingTriggerAbilityFromDef(pendingTriggerSourceDef(g, trigger), trigger)
}

func pendingTriggerSourceDef(g *game.Game, trigger pendingTriggeredAbility) *game.CardDef {
	var def *game.CardDef
	if trigger.sourceCardID != 0 {
		if card := g.GetCardInstance(trigger.sourceCardID); card != nil {
			def = card.Def
		}
	} else {
		def = trigger.sourceToken
	}
	return def
}

func pendingTriggerAbilityFromDef(def *game.CardDef, trigger pendingTriggeredAbility) *game.AbilityDef {
	if def == nil || trigger.abilityIndex < 0 || trigger.abilityIndex >= len(def.Abilities) {
		return nil
	}
	return &def.Abilities[trigger.abilityIndex]
}

func (e *Engine) chooseTriggerOrder(g *game.Game, playerID game.PlayerID, triggers []pendingTriggeredAbility, agents [game.NumPlayers]PlayerAgent, log *TurnLog) []pendingTriggeredAbility {
	if len(triggers) <= 1 {
		return triggers
	}
	options := make([]game.ChoiceOption, 0, len(triggers))
	for i, trigger := range triggers {
		options = append(options, game.ChoiceOption{
			Index: i,
			Label: fmt.Sprintf("source=%v ability=%d", trigger.sourceID, trigger.abilityIndex),
		})
	}
	selected := e.chooseChoice(g, agents, orderChoiceRequest(playerID, "Order triggered abilities.", options), log)
	ordered := make([]pendingTriggeredAbility, 0, len(triggers))
	used := make([]bool, len(triggers))
	for _, index := range selected {
		if index < 0 || index >= len(triggers) || used[index] {
			continue
		}
		ordered = append(ordered, triggers[index])
		used[index] = true
	}
	for i, trigger := range triggers {
		if !used[i] {
			ordered = append(ordered, trigger)
		}
	}
	return ordered
}

func triggerAPNAPPlayers(g *game.Game) []game.PlayerID {
	if g == nil {
		return nil
	}
	players := make([]game.PlayerID, 0, game.NumPlayers)
	playerID := g.Turn.ActivePlayer
	for range int(game.NumPlayers) {
		if playerID < 0 || playerID >= game.NumPlayers {
			break
		}
		players = append(players, playerID)
		playerID = g.TurnOrder.NextPriority(playerID)
		if slices.Contains(players, playerID) {
			break
		}
	}
	return players
}
