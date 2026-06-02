package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

type pendingTriggeredAbility struct {
	controller   game.PlayerID
	sourceID     id.ID
	sourceCardID id.ID
	sourceToken  *game.CardDef
	face         game.FaceIndex
	abilityIndex int
	targets      []game.Target
	event        game.GameEvent
	hasEvent     bool
	inline       *game.AbilityDef
	wardTargetID id.ID
}

func (e *Engine) putTriggeredAbilitiesOnStack(g *game.Game) bool {
	return e.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, nil)
}

func (e *Engine) putTriggeredAbilitiesOnStackWithChoices(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	start := g.TriggerEventCursor
	if start < 0 || start > len(g.Events) {
		start = len(g.Events)
	}
	events := append([]game.GameEvent(nil), g.Events[start:]...)
	g.TriggerEventCursor = len(g.Events)
	pending := e.detectTriggeredAbilities(g, events)
	pending = append(pending, e.detectMadnessTriggeredAbilities(g, events)...)
	pending = append(pending, e.detectStateTriggeredAbilities(g)...)
	if len(pending) == 0 {
		return false
	}
	orderedTriggers := e.orderTriggeredAbilitiesAPNAP(g, pending, agents, log)
	for i := range orderedTriggers {
		trigger := &orderedTriggers[i]
		obj := &game.StackObject{
			ID:                      g.IDGen.Next(),
			Kind:                    game.StackTriggeredAbility,
			SourceID:                trigger.sourceID,
			Face:                    trigger.face,
			SourceCardID:            trigger.sourceCardID,
			SourceTokenDef:          trigger.sourceToken,
			AbilityIndex:            trigger.abilityIndex,
			TriggerEvent:            trigger.event,
			HasTriggerEvent:         trigger.hasEvent,
			InlineAbility:           trigger.inline,
			WardTargetStackObjectID: trigger.wardTargetID,
			Controller:              trigger.controller,
			Targets:                 append([]game.Target(nil), trigger.targets...),
		}
		pushAbilityToStack(g, obj)
	}
	return true
}

func (*Engine) detectMadnessTriggeredAbilities(g *game.Game, events []game.GameEvent) []pendingTriggeredAbility {
	var pending []pendingTriggeredAbility
	for _, event := range events {
		if event.Kind != game.EventCardDiscarded || event.ToZone != game.ZoneExile || event.CardID == 0 {
			continue
		}
		card, ok := g.GetCardInstance(event.CardID)
		if !ok {
			continue
		}
		cost, ok := madnessCostForCard(cardFaceOrDefault(card, event.Face))
		if !ok {
			continue
		}
		pending = append(pending, pendingTriggeredAbility{
			controller:   event.Player,
			sourceID:     event.CardID,
			sourceCardID: event.CardID,
			face:         game.FaceFront,
			inline: &game.AbilityDef{
				Kind:        game.TriggeredAbility,
				Text:        "Madness",
				Keywords:    []game.Keyword{game.Madness},
				MadnessCost: opt.Val(cost),
			},
			event:    event,
			hasEvent: true,
		})
	}
	return pending
}

func (e *Engine) detectTriggeredAbilities(g *game.Game, events []game.GameEvent) []pendingTriggeredAbility {
	var pending []pendingTriggeredAbility
	for _, event := range events {
		for _, permanent := range g.Battlefield {
			pending = append(pending, e.detectTriggeredAbilitiesFromPermanent(g, permanent, event)...)
		}
		if source, ok := leftBattlefieldTriggerSource(g, event); ok {
			pending = append(pending, e.detectTriggeredAbilitiesFromPermanent(g, source, event)...)
		}
	}
	return pending
}

func (*Engine) detectTriggeredAbilitiesFromPermanent(g *game.Game, permanent *game.Permanent, event game.GameEvent) []pendingTriggeredAbility {
	abilities := permanentEffectiveAbilities(g, permanent)
	var pending []pendingTriggeredAbility
	controller := effectiveController(g, permanent)
	for i := range abilities {
		ability := &abilities[i]
		if ability.Kind != game.TriggeredAbility || !ability.Trigger.Exists {
			if ward, ok := wardTriggerForEvent(g, permanent, controller, ability, event); ok {
				pending = append(pending, pendingTriggeredAbility{
					controller:   controller,
					sourceID:     permanent.ObjectID,
					sourceCardID: permanent.CardInstanceID,
					sourceToken:  permanent.TokenDef,
					face:         permanent.Face,
					inline:       ward,
					event:        event,
					hasEvent:     true,
					wardTargetID: event.StackObjectID,
				})
			}
			continue
		}
		trigger := &ability.Trigger.Val
		if !triggerMatchesEvent(g, permanent, trigger.Pattern, event) || !triggerInterveningIf(g, permanent, controller, trigger, &event) {
			continue
		}
		pending = append(pending, pendingTriggeredAbility{
			controller:   controller,
			sourceID:     permanent.ObjectID,
			sourceCardID: permanent.CardInstanceID,
			sourceToken:  permanent.TokenDef,
			face:         permanent.Face,
			abilityIndex: i,
			event:        event,
			hasEvent:     true,
		})
	}
	if prowess, ok := prowessTriggerForEvent(g, permanent, controller, event); ok {
		pending = append(pending, pendingTriggeredAbility{
			controller:   controller,
			sourceID:     permanent.ObjectID,
			sourceCardID: permanent.CardInstanceID,
			sourceToken:  permanent.TokenDef,
			face:         permanent.Face,
			inline:       prowess,
			event:        event,
			hasEvent:     true,
		})
	}
	return pending
}

func wardTriggerForEvent(g *game.Game, permanent *game.Permanent, controller game.PlayerID, ability *game.AbilityDef, event game.GameEvent) (*game.AbilityDef, bool) {
	if event.Kind != game.EventObjectBecameTarget || event.PermanentID != permanent.ObjectID || event.StackObjectID == 0 {
		return nil, false
	}
	if event.Controller == controller || !abilityHasKeyword(ability, game.Ward) || !ability.WardCost.Exists {
		return nil, false
	}
	return &game.AbilityDef{
		Kind:     game.TriggeredAbility,
		Text:     "Ward",
		Keywords: []game.Keyword{game.Ward},
		WardCost: ability.WardCost,
	}, true
}

func prowessTriggerForEvent(g *game.Game, permanent *game.Permanent, controller game.PlayerID, event game.GameEvent) (*game.AbilityDef, bool) {
	if event.Kind != game.EventSpellCast || event.Controller != controller || !hasKeyword(g, permanent, game.Prowess) {
		return nil, false
	}
	if slices.Contains(event.CardTypes, types.Creature) {
		return nil, false
	}
	return &game.AbilityDef{
		Kind: game.TriggeredAbility,
		Text: "Prowess",
		Effects: []game.Effect{
			{
				Type:           game.EffectModifyPT,
				TargetIndex:    -2,
				PowerDelta:     1,
				ToughnessDelta: 1,
				UntilEndOfTurn: true,
			},
		},
	}, true
}

func (*Engine) detectStateTriggeredAbilities(g *game.Game) []pendingTriggeredAbility {
	if g.StateTriggerLatches == nil {
		g.StateTriggerLatches = make(map[game.StateTriggerKey]bool)
	}
	var pending []pendingTriggeredAbility
	seen := make(map[game.StateTriggerKey]bool)
	for _, permanent := range g.Battlefield {
		def, ok := permanentCardDef(g, permanent)
		if !ok {
			continue
		}
		controller := effectiveController(g, permanent)
		for i := range def.Abilities {
			ability := &def.Abilities[i]
			if ability.Kind != game.TriggeredAbility || !ability.Trigger.Exists || !ability.Trigger.Val.State.Exists {
				continue
			}
			key := game.StateTriggerKey{
				SourceObjectID: permanent.ObjectID,
				SourceCardID:   permanent.CardInstanceID,
				AbilityIndex:   i,
			}
			seen[key] = true
			if !stateTriggerConditionSatisfied(g, controller, &ability.Trigger.Val.State.Val) {
				delete(g.StateTriggerLatches, key)
				continue
			}
			if g.StateTriggerLatches[key] {
				continue
			}
			// State triggers fire once while their condition is true and do not
			// trigger again until the condition becomes false, then true (CR 603.8).
			g.StateTriggerLatches[key] = true
			pending = append(pending, pendingTriggeredAbility{
				controller:   controller,
				sourceID:     permanent.ObjectID,
				sourceCardID: permanent.CardInstanceID,
				sourceToken:  permanent.TokenDef,
				face:         permanent.Face,
				abilityIndex: i,
			})
		}
	}
	for key := range g.StateTriggerLatches {
		if !seen[key] {
			delete(g.StateTriggerLatches, key)
		}
	}
	return pending
}

func stateTriggerConditionSatisfied(g *game.Game, controller game.PlayerID, condition *game.StateTriggerCondition) bool {
	if condition == nil {
		return false
	}
	if condition.MatchControllerLifeLessOrEqual {
		player, ok := playerByID(g, controller)
		if !ok || player.Life > condition.ControllerLifeLessOrEqual {
			return false
		}
	}
	return true
}

func leftBattlefieldTriggerSource(g *game.Game, event game.GameEvent) (*game.Permanent, bool) {
	if event.FromZone != game.ZoneBattlefield || event.PermanentID == 0 {
		return nil, false
	}
	if _, ok := permanentByObjectID(g, event.PermanentID); ok {
		return nil, false
	}
	if event.CardID != 0 {
		if _, ok := g.GetCardInstance(event.CardID); !ok {
			return nil, false
		}
		return &game.Permanent{
			ObjectID:       event.PermanentID,
			CardInstanceID: event.CardID,
			Owner:          event.Player,
			Controller:     event.Controller,
			Face:           event.Face,
		}, true
	}
	if event.TokenDef == nil {
		return nil, false
	}
	return &game.Permanent{
		ObjectID:   event.PermanentID,
		Owner:      event.Player,
		Controller: event.Controller,
		Face:       event.Face,
		Token:      true,
		TokenDef:   event.TokenDef,
	}, true
}

func (e *Engine) triggerTargets(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, ability *game.AbilityDef, agents [game.NumPlayers]PlayerAgent, log *TurnLog) ([]game.Target, bool) {
	result := targetChoicesForAbilityFromSourceObject(g, controller, source, sourceObjectID, ability)
	switch result.kind {
	case targetNoLegalChoices, targetInvalidSpec:
		return nil, false
	}
	choices := result.choices
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
	if pattern.Event == game.EventUnknown || pattern.Event != event.Kind {
		return false
	}

	// Trigger patterns are checked when the triggering event is processed, and
	// LTB/dies checks may need last-known information for the moved permanent
	// (CR 603.2, CR 603.6c, CR 603.10).
	sourceController := effectiveController(g, source)
	if !triggerControllerMatches(sourceController, pattern.Controller, event.Controller) {
		return false
	}
	if !triggerSourceMatches(g, source, pattern.Source, event) {
		return false
	}
	if pattern.ExcludeSelf && triggerSourceMatches(g, source, game.TriggerSourceSelf, event) {
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
	if pattern.Event == game.EventBeginningOfStep {
		if pattern.Step == game.StepNone || pattern.Step != event.Step {
			return false
		}
	}
	if !eventPermanentTypeFiltersMatch(g, event, pattern.RequirePermanentTypes, pattern.ExcludePermanentTypes) {
		return false
	}
	if pattern.RequireNonToken && eventPermanentIsToken(g, event) {
		return false
	}
	if !eventCardTypeFiltersMatch(g, event, pattern.RequireCardTypes, pattern.ExcludeCardTypes) {
		return false
	}
	return true
}

func eventPermanentIsToken(g *game.Game, event game.GameEvent) bool {
	if event.PermanentID != 0 {
		if permanent, ok := permanentByObjectID(g, event.PermanentID); ok {
			return permanent.Token
		}
		if snapshot, ok := lastKnownObject(g, event.PermanentID); ok {
			return snapshot.TokenDef != nil || snapshot.CardID == 0
		}
	}
	return event.TokenDef != nil || (event.CardID == 0 && event.TokenName != "")
}

func triggerInterveningIf(g *game.Game, source *game.Permanent, controller game.PlayerID, trigger *game.TriggerCondition, event *game.GameEvent) bool {
	if trigger == nil {
		return true
	}
	// Intervening "if" conditions are checked both as the event triggers and as
	// the ability resolves (CR 603.4).
	if trigger.InterveningIfControllerLifeAtLeast != 0 {
		player, ok := playerByID(g, controller)
		if !ok || player.Life < trigger.InterveningIfControllerLifeAtLeast {
			return false
		}
	}
	if trigger.InterveningIfEventPermanentHadCounters && !eventPermanentHadCounters(g, event) {
		return false
	}
	if !conditionSatisfied(g, conditionContext{
		controller: controller,
		source:     source,
		event:      event,
	}, trigger.InterveningCondition) {
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

func triggerSourceMatches(g *game.Game, source *game.Permanent, filter game.TriggerSourceFilter, event game.GameEvent) bool {
	if filter == game.TriggerSourceAttachedPermanent {
		return triggerSourceAttachedPermanentMatches(g, source, event)
	}
	if filter != game.TriggerSourceSelf {
		return true
	}
	return (source.ObjectID != 0 && event.SourceObjectID == source.ObjectID) ||
		(source.ObjectID != 0 && event.PermanentID == source.ObjectID) ||
		(source.CardInstanceID != 0 && event.SourceID == source.CardInstanceID) ||
		(source.CardInstanceID != 0 && event.CardID == source.CardInstanceID)
}

func triggerSourceAttachedPermanentMatches(g *game.Game, source *game.Permanent, event game.GameEvent) bool {
	if source.ObjectID == 0 || event.PermanentID == 0 {
		return false
	}
	if source.AttachedTo.Exists && source.AttachedTo.Val == event.PermanentID {
		return true
	}
	if snapshot, ok := lastKnownObject(g, event.PermanentID); ok {
		return slices.Contains(snapshot.Attachments, source.ObjectID)
	}
	return false
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

func eventPermanentHasType(g *game.Game, event game.GameEvent, cardType types.Card) bool {
	if event.PermanentID != 0 {
		if permanent, ok := permanentByObjectID(g, event.PermanentID); ok {
			return permanentHasType(g, permanent, cardType)
		}
		// Leaves-the-battlefield and dies triggers look back at the permanent's
		// last existence on the battlefield (CR 603.10).
		if snapshot, ok := lastKnownObject(g, event.PermanentID); ok {
			return slices.Contains(snapshot.Types, cardType)
		}
	}
	if event.CardID != 0 {
		if card, ok := g.GetCardInstance(event.CardID); ok {
			return cardFaceOrDefault(card, game.FaceFront).HasType(cardType)
		}
	}
	if event.TokenDef != nil {
		return event.TokenDef.HasType(cardType)
	}
	return false
}

func eventPermanentTypeFiltersMatch(g *game.Game, event game.GameEvent, required, excluded []types.Card) bool {
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

func eventCardTypeFiltersMatch(g *game.Game, event game.GameEvent, required, excluded []types.Card) bool {
	cardTypes := event.CardTypes
	if len(cardTypes) == 0 && event.CardID != 0 {
		if card, ok := g.GetCardInstance(event.CardID); ok {
			cardTypes = cardFaceOrDefault(card, game.FaceFront).Types
		}
	}
	for _, cardType := range required {
		if !slices.Contains(cardTypes, cardType) {
			return false
		}
	}
	for _, cardType := range excluded {
		if slices.Contains(cardTypes, cardType) {
			return false
		}
	}
	return true
}

func eventPermanentHadCounters(g *game.Game, event *game.GameEvent) bool {
	if event == nil || event.PermanentID == 0 {
		return false
	}
	if permanent, ok := permanentByObjectID(g, event.PermanentID); ok {
		return !permanent.Counters.IsEmpty()
	}
	if snapshot, ok := lastKnownObject(g, event.PermanentID); ok {
		return !snapshot.Counters.IsEmpty()
	}
	return false
}

func (e *Engine) orderTriggeredAbilitiesAPNAP(g *game.Game, triggers []pendingTriggeredAbility, agents [game.NumPlayers]PlayerAgent, log *TurnLog) []pendingTriggeredAbility {
	if len(triggers) == 0 {
		return triggers
	}
	ordered := make([]pendingTriggeredAbility, 0, len(triggers))
	used := make([]bool, len(triggers))
	for _, playerID := range triggerAPNAPPlayers(g) {
		var playerTriggers []pendingTriggeredAbility
		for i := range triggers {
			trigger := &triggers[i]
			if trigger.controller == playerID {
				playerTriggers = append(playerTriggers, *trigger)
				used[i] = true
			}
		}
		ordered = append(ordered, e.preparePlayerTriggers(g, playerID, playerTriggers, agents, log)...)
	}
	for i := range triggers {
		if !used[i] {
			ordered = append(ordered, triggers[i])
		}
	}
	return ordered
}

func (e *Engine) preparePlayerTriggers(g *game.Game, playerID game.PlayerID, triggers []pendingTriggeredAbility, agents [game.NumPlayers]PlayerAgent, log *TurnLog) []pendingTriggeredAbility {
	ordered := e.chooseTriggerOrder(g, playerID, triggers, agents, log)
	prepared := make([]pendingTriggeredAbility, 0, len(ordered))
	for i := range ordered {
		trigger := &ordered[i]
		source, _ := pendingTriggerSourceDef(g, trigger)
		ability, ok := pendingTriggerAbilityFromDef(source, trigger)
		if !ok {
			continue
		}
		targets, ok := e.triggerTargets(g, trigger.controller, source, trigger.sourceID, ability, agents, log)
		if !ok {
			continue
		}
		trigger.targets = targets
		prepared = append(prepared, *trigger)
	}
	return prepared
}

func pendingTriggerAbility(g *game.Game, trigger *pendingTriggeredAbility) (*game.AbilityDef, bool) {
	source, _ := pendingTriggerSourceDef(g, trigger)
	return pendingTriggerAbilityFromDef(source, trigger)
}

func pendingTriggerSourceDef(g *game.Game, trigger *pendingTriggeredAbility) (*game.CardDef, bool) {
	if trigger.sourceCardID != 0 {
		if card, ok := g.GetCardInstance(trigger.sourceCardID); ok {
			return card.Def.FaceDef(trigger.face)
		}
		return nil, false
	}
	if trigger.sourceToken == nil {
		return nil, false
	}
	return trigger.sourceToken.FaceDef(trigger.face)
}

func pendingTriggerAbilityFromDef(def *game.CardDef, trigger *pendingTriggeredAbility) (*game.AbilityDef, bool) {
	if trigger.inline != nil {
		return trigger.inline, true
	}
	if def == nil || trigger.abilityIndex < 0 || trigger.abilityIndex >= len(def.Abilities) {
		return nil, false
	}
	return &def.Abilities[trigger.abilityIndex], true
}

func (e *Engine) chooseTriggerOrder(g *game.Game, playerID game.PlayerID, triggers []pendingTriggeredAbility, agents [game.NumPlayers]PlayerAgent, log *TurnLog) []pendingTriggeredAbility {
	if len(triggers) <= 1 {
		return triggers
	}
	options := make([]game.ChoiceOption, 0, len(triggers))
	for i := range triggers {
		trigger := &triggers[i]
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
	for i := range triggers {
		if !used[i] {
			ordered = append(ordered, triggers[i])
		}
	}
	return ordered
}

func triggerAPNAPPlayers(g *game.Game) []game.PlayerID {
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
