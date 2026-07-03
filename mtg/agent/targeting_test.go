package agent

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/rules"
)

func TestChooseTargetPicksBiggestOpponentThreat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	small := addObservedPermanent(g, game.Player2, creatureCardDef("Rat", 2, 2))
	big := addObservedPermanent(g, game.Player2, creatureCardDef("Wurm", 6, 6))
	obs := rules.NewObservation(g, game.Player1)

	request := game.ChoiceRequest{
		Kind:   game.ChoiceTarget,
		Player: game.Player1,
		Options: []game.ChoiceOption{
			{Index: 0, Targets: []game.Target{game.PermanentTarget(small.ObjectID)}},
			{Index: 1, Targets: []game.Target{game.PermanentTarget(big.ObjectID)}},
		},
		MinChoices: 1,
		MaxChoices: 1,
	}

	got := GenericStrategy{}.ChooseChoice(obs, request)
	if len(got) != 1 || got[0] != 1 {
		t.Fatalf("target choice = %v, want option 1 (the 6/6, the biggest threat)", got)
	}
}

func TestChooseTargetBuffsBiggestOwnWhenNoOpponentOption(t *testing.T) {
	// A "target creature you control" choice offers only the agent's own
	// creatures, so it should land on the biggest one (a buff on the best body).
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	small := addObservedPermanent(g, game.Player1, creatureCardDef("Squire", 1, 1))
	big := addObservedPermanent(g, game.Player1, creatureCardDef("Knight", 5, 5))
	obs := rules.NewObservation(g, game.Player1)

	request := game.ChoiceRequest{
		Kind:   game.ChoiceTarget,
		Player: game.Player1,
		Options: []game.ChoiceOption{
			{Index: 0, Targets: []game.Target{game.PermanentTarget(small.ObjectID)}},
			{Index: 1, Targets: []game.Target{game.PermanentTarget(big.ObjectID)}},
		},
		MinChoices: 1,
		MaxChoices: 1,
	}

	got := GenericStrategy{}.ChooseChoice(obs, request)
	if len(got) != 1 || got[0] != 1 {
		t.Fatalf("target choice = %v, want option 1 (the 5/5, the best own body)", got)
	}
}

func TestChoosePlayerPicksBiggestThreatOpponent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedPermanent(g, game.Player2, creatureCardDef("Wurm", 8, 8))
	// Player3 has no board, so Player2 is the bigger threat.
	obs := rules.NewObservation(g, game.Player1)

	request := game.ChoiceRequest{
		Kind:   game.ChoicePlayer,
		Player: game.Player1,
		Options: []game.ChoiceOption{
			{Index: 0, Targets: []game.Target{game.PlayerTarget(game.Player3)}},
			{Index: 1, Targets: []game.Target{game.PlayerTarget(game.Player2)}},
		},
		MinChoices: 1,
		MaxChoices: 1,
	}

	got := GenericStrategy{}.ChooseChoice(obs, request)
	if len(got) != 1 || got[0] != 1 {
		t.Fatalf("player choice = %v, want option 1 (Player2, the bigger threat)", got)
	}
}

func TestChooseTargetFallsBackWithoutResolvableTargets(t *testing.T) {
	// Options with no target references fall back to the baseline default so the
	// engine's validation and fallback still apply.
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obs := rules.NewObservation(g, game.Player1)
	request := game.ChoiceRequest{
		Kind:             game.ChoiceTarget,
		Player:           game.Player1,
		Options:          []game.ChoiceOption{{Index: 0, Label: "A"}, {Index: 1, Label: "B"}},
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}
	got := GenericStrategy{}.ChooseChoice(obs, request)
	if len(got) != 1 || got[0] != 0 {
		t.Fatalf("fallback choice = %v, want the default [0]", got)
	}
}
