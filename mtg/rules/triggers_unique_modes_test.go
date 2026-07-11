package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestTriggeredModesUniquePerTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	ability := &game.TriggeredAbility{Content: game.AbilityContent{
		MinModes:           1,
		MaxModes:           1,
		ModesUniquePerTurn: true,
		Modes: []game.Mode{
			{Text: "First"},
			{Text: "Second"},
			{Text: "Third"},
		},
	}}
	use := game.TriggeredAbilityUse{SourceID: 101, AbilityIndex: 2}
	agent := &choiceOnlyAgent{choices: [][]int{{0}, {1}, {2}}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	log := &TurnLog{}

	for want := range 3 {
		selected, ok := engine.triggerModes(g, game.Player1, use, ability, agents, log)
		if !ok || !slices.Equal(selected, []int{want}) {
			t.Fatalf("choice %d = %v/%v, want [%d]/true", want, selected, ok, want)
		}
		request := log.Choices[len(log.Choices)-1].Request
		if len(request.Options) != 3-want || request.Options[0].Index != want {
			t.Fatalf("choice %d options = %#v, want remaining printed modes", want, request.Options)
		}
	}
	if _, ok := engine.triggerModes(g, game.Player1, use, ability, agents, log); ok {
		t.Fatal("fourth choice succeeded after every mode was chosen")
	}
	if len(log.Choices) != 3 {
		t.Fatalf("choice requests = %d, want no prompt for exhausted modes", len(log.Choices))
	}

	otherUse := game.TriggeredAbilityUse{SourceID: 102, AbilityIndex: 2}
	otherAgent := &choiceOnlyAgent{choices: [][]int{{0}}}
	otherAgents := [game.NumPlayers]PlayerAgent{game.Player1: otherAgent}
	if selected, ok := engine.triggerModes(g, game.Player1, otherUse, ability, otherAgents, &TurnLog{}); !ok ||
		!slices.Equal(selected, []int{0}) {
		t.Fatalf("other source choice = %v/%v, want independent [0]/true", selected, ok)
	}

	clone := g.Clone()
	g.ChosenModesThisTurn[use] = 0
	if clone.ChosenModesThisTurn[use] != 0b111 {
		t.Fatalf("cloned mode history = %b, want independent 111", clone.ChosenModesThisTurn[use])
	}

	engine.advanceToNextTurn(g)
	resetAgent := &choiceOnlyAgent{choices: [][]int{{0}}}
	resetAgents := [game.NumPlayers]PlayerAgent{game.Player1: resetAgent}
	if selected, ok := engine.triggerModes(g, game.Player1, use, ability, resetAgents, &TurnLog{}); !ok ||
		!slices.Equal(selected, []int{0}) {
		t.Fatalf("new-turn choice = %v/%v, want reset [0]/true", selected, ok)
	}
}
