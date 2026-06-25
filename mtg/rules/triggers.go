package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
)

// orderTriggeredAbilitiesAPNAP orders pending triggered abilities for placement
// on the stack in APNAP order (CR 603.3b): each player, in turn order starting
// with the active player (CR 101.4), puts the triggers they control on the stack
// in the order they choose. CR 603.3b technically describes two APNAP passes
// (abilities that trigger on another ability triggering are placed in the second
// pass); this engine uses a single APNAP pass, which matches the rules except
// for the rare ability that triggers on a triggered ability being put on the
// stack. Because the stack is last-in-first-out, the last player's triggers
// resolve first. Triggers with no identifiable controller are appended last.
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

// prepareTriggeredAbility makes the choices required as a triggered ability is
// put on the stack: its controller announces modes (CR 603.3c) and chooses
// targets (CR 603.3d, which defers to the casting choices of CR 601.2c-d). It
// reports false when the ability can make no legal choice (e.g. no legal mode or
// no legal target), in which case the ability is removed from the stack
// (CR 603.3c, CR 603.3d).
func (e *Engine) prepareTriggeredAbility(g *game.Game, trigger *pendingTriggeredAbility, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	source, _ := pendingTriggerSourceDef(g, trigger)
	ability, ok := pendingTriggerAbilityFromDef(source, trigger)
	if !ok {
		return false
	}
	chosenModes, ok := e.triggerModes(g, trigger.controller, ability, agents, log)
	if !ok {
		return false
	}
	trigger.chosenModes = chosenModes
	targets, ok := e.triggerTargets(g, trigger.controller, source, trigger.sourceID, ability, chosenModes, agents, log)
	if !ok {
		return false
	}
	trigger.targets = targets
	targetCounts, ok := bodyTargetCountsWithModes(g, trigger.controller, source, trigger.sourceID, ability, chosenModes, targets)
	if !ok {
		panic("validated triggered ability targets could not be segmented")
	}
	trigger.targetCounts = targetCounts
	return true
}

// triggerModes selects the mode(s) for a modal triggered ability as it is put on
// the stack (CR 603.3c: a modal triggered ability's controller announces the
// mode choice; an illegal mode can't be chosen; if no mode is chosen the ability
// is removed from the stack). It returns ok=false to signal that no legal mode
// could be chosen so the caller removes the ability from the stack.
func (e *Engine) triggerModes(g *game.Game, controller game.PlayerID, ability *game.TriggeredAbility, agents [game.NumPlayers]PlayerAgent, log *TurnLog) ([]int, bool) {
	content := ability.Content
	if !content.IsModal() {
		return nil, true
	}
	if content.AllowDuplicateModes {
		return nil, false
	}
	if len(content.SharedTargets) != 0 {
		return nil, false
	}
	minModes, maxModes := modeChoiceRangeFromContent(content)
	if !modeChoiceRangeValid(content, minModes, maxModes) {
		return nil, false
	}
	if content.RandomModes {
		// "Choose one at random" selects the single mode with the game's random
		// source rather than prompting the controller (CR 700.2). The lowering
		// guarantees a one/one range, so exactly one mode is selected.
		if minModes != 1 || maxModes != 1 || len(content.Modes) == 0 {
			return nil, false
		}
		selected := []int{e.rng.IntN(len(content.Modes))}
		if !modesValidForContent(content, selected) {
			return nil, false
		}
		return selected, true
	}
	options := make([]game.ChoiceOption, len(content.Modes))
	for i := range content.Modes {
		options[i] = game.ChoiceOption{Index: i, Label: content.Modes[i].Text}
	}
	defaultSelection := make([]int, minModes)
	for i := range defaultSelection {
		defaultSelection[i] = i
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:             game.ChoiceModal,
		Player:           controller,
		Prompt:           "Choose modes for triggered ability.",
		Options:          options,
		MinChoices:       minModes,
		MaxChoices:       maxModes,
		DefaultSelection: defaultSelection,
	}, log)
	slices.Sort(selected)
	if !modesValidForContent(content, selected) {
		return nil, false
	}
	return selected, true
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

// chooseTriggerOrder lets a single player order the triggered abilities they
// control before they are placed on the stack (CR 603.3b: "in any order they
// choose"). Stack order is last-in-first-out, so the ability placed last
// resolves first.
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

// triggerAPNAPPlayers returns the players in APNAP order (CR 101.4): the active
// player first, then each other player in turn order. This ordering governs how
// simultaneous triggered abilities are placed on the stack (CR 603.3b).
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
