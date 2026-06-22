package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

// tapOrUntapChoiceAgent answers the "Tap or untap the permanent?" choice by
// returning the configured option index.
type tapOrUntapChoiceAgent struct {
	choice int
}

func (tapOrUntapChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a tapOrUntapChoiceAgent) ChooseChoice(_ PlayerObservation, _ game.ChoiceRequest) []int {
	return []int{a.choice}
}

func TestTapOrUntapEffectTapsWhenTapChosen(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, game.TapOrUntap{Object: game.TargetPermanentReference(0)}, []game.Target{game.PermanentTarget(target.ObjectID)})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: tapOrUntapChoiceAgent{choice: 0}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !target.Tapped {
		t.Fatal("permanent was not tapped after choosing Tap")
	}
}

func TestTapOrUntapEffectUntapsWhenUntapChosen(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCreaturePermanent(g, game.Player2)
	target.Tapped = true
	addEffectSpellToStack(g, game.Player1, game.TapOrUntap{Object: game.TargetPermanentReference(0)}, []game.Target{game.PermanentTarget(target.ObjectID)})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: tapOrUntapChoiceAgent{choice: 1}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if target.Tapped {
		t.Fatal("permanent remained tapped after choosing Untap")
	}
}
