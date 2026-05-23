package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

type chooseActionAgent struct {
	chosen action.Action
}

func (a chooseActionAgent) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	return a.chosen
}

func TestRunPriorityLoopNilAgentsPassAroundTable(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.runPriorityLoop(g, [game.NumPlayers]PlayerAgent{}, &log)

	if len(log.Actions) != game.NumPlayers {
		t.Fatalf("logged actions = %d, want %d", len(log.Actions), game.NumPlayers)
	}
	for i, logged := range log.Actions {
		if logged.Player != game.PlayerID(i) {
			t.Fatalf("logged action %d player = %v, want %v", i, logged.Player, game.PlayerID(i))
		}
		if logged.Action.Kind != action.ActionPass {
			t.Fatalf("logged action %d kind = %v, want %v", i, logged.Action.Kind, action.ActionPass)
		}
	}
}

func TestRunPriorityLoopInvalidAgentActionFallsBackToPass(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	log := TurnLog{}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: chooseActionAgent{chosen: action.PlayLand(42)},
	}

	engine.runPriorityLoop(g, agents, &log)

	if len(log.Actions) != game.NumPlayers {
		t.Fatalf("logged actions = %d, want %d", len(log.Actions), game.NumPlayers)
	}
	if log.Actions[0].Action.Kind != action.ActionPass {
		t.Fatalf("logged action kind = %v, want %v", log.Actions[0].Action.Kind, action.ActionPass)
	}
}

func TestRunPriorityLoopSkipsPriorityPlayerEliminatedByStateBasedActions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.FailedDraws[game.Player1] = true
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.runPriorityLoop(g, [game.NumPlayers]PlayerAgent{}, &log)

	if len(log.Actions) != game.NumPlayers-1 {
		t.Fatalf("logged actions = %d, want %d", len(log.Actions), game.NumPlayers-1)
	}
	for _, logged := range log.Actions {
		if logged.Player == game.Player1 {
			t.Fatal("eliminated player was asked to act")
		}
	}
}
