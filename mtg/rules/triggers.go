package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
)

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
		ordered = append(ordered, e.chooseTriggerOrder(g, playerID, playerTriggers, agents, log)...)
	}
	for i := range triggers {
		if !used[i] {
			ordered = append(ordered, triggers[i])
		}
	}
	return ordered
}

func (e *Engine) prepareTriggeredAbility(g *game.Game, trigger *pendingTriggeredAbility, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	source, _ := pendingTriggerSourceDef(g, trigger)
	ability, ok := pendingTriggerAbilityFromDef(source, trigger)
	if !ok {
		return false
	}
	targets, ok := e.triggerTargets(g, trigger.controller, source, trigger.sourceID, ability, agents, log)
	if !ok {
		return false
	}
	trigger.targets = targets
	targetCounts, ok := bodyTargetCounts(g, trigger.controller, source, trigger.sourceID, ability, targets)
	if !ok {
		panic("validated triggered ability targets could not be segmented")
	}
	trigger.targetCounts = targetCounts
	return true
}

func releasePendingStateTriggerLatch(g *game.Game, trigger *pendingTriggeredAbility) {
	if trigger.inline == nil || !trigger.inline.Trigger.State.Exists {
		return
	}
	deleteStateTriggerLatch(g, trigger.sourceID, trigger.sourceCardID, trigger.abilityIndex)
}

func pendingTriggerAbility(g *game.Game, trigger *pendingTriggeredAbility) (*game.TriggeredAbility, bool) {
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

func pendingTriggerAbilityFromDef(def *game.CardDef, trigger *pendingTriggeredAbility) (*game.TriggeredAbility, bool) {
	if trigger.inline != nil {
		return trigger.inline, true
	}
	if def == nil {
		return nil, false
	}
	body := def.BodyAt(trigger.abilityIndex)
	triggered, ok := body.(*game.TriggeredAbility)
	if !ok {
		return nil, false
	}
	return triggered, true
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
