package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
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

func TestRunPriorityLoopAllPassWithNonEmptyStackResolvesAndContinues(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Stack.Push(&game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		Controller: game.Player1,
	})
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.runPriorityLoop(g, [game.NumPlayers]PlayerAgent{}, &log)

	if !g.Stack.IsEmpty() {
		t.Fatal("stack is not empty after all players passed")
	}
	wantActions := game.NumPlayers * 2
	if len(log.Actions) != wantActions {
		t.Fatalf("logged actions = %d, want %d", len(log.Actions), wantActions)
	}
	if log.Actions[game.NumPlayers].Player != g.Turn.ActivePlayer {
		t.Fatalf("first priority after resolution = %v, want active player %v", log.Actions[game.NumPlayers].Player, g.Turn.ActivePlayer)
	}
}

func TestRunPriorityLoopNonPassActionKeepsPriorityWithActor(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	landID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:  "Forest",
		Types: []types.Card{types.Land},
	})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	log := TurnLog{}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: chooseActionAgent{chosen: action.PlayLand(landID)},
	}

	engine.runPriorityLoop(g, agents, &log)

	if len(log.Actions) != game.NumPlayers+1 {
		t.Fatalf("logged actions = %d, want %d", len(log.Actions), game.NumPlayers+1)
	}
	if !actionsEqual(log.Actions[0].Action, action.PlayLand(landID)) {
		t.Fatalf("first action = %+v, want PlayLand(%v)", log.Actions[0].Action, landID)
	}
	if log.Actions[1].Player != game.Player1 {
		t.Fatalf("priority after non-pass action = %v, want %v", log.Actions[1].Player, game.Player1)
	}
}

func TestActionsEqualDistinguishesTargetKindAndValue(t *testing.T) {
	cardID := id.ID(42)
	playerTarget := game.PlayerTarget(game.Player2)
	otherPlayerTarget := game.PlayerTarget(game.Player3)
	permanentTarget := game.PermanentTarget(id.ID(game.Player2))

	if !actionsEqual(action.CastSpell(cardID, []game.Target{playerTarget}, 0, nil), action.CastSpell(cardID, []game.Target{playerTarget}, 0, nil)) {
		t.Fatal("actionsEqual() = false for matching player target")
	}
	if actionsEqual(action.CastSpell(cardID, []game.Target{playerTarget}, 0, nil), action.CastSpell(cardID, []game.Target{otherPlayerTarget}, 0, nil)) {
		t.Fatal("actionsEqual() = true for different player target values")
	}
	if actionsEqual(action.CastSpell(cardID, []game.Target{playerTarget}, 0, nil), action.CastSpell(cardID, []game.Target{permanentTarget}, 0, nil)) {
		t.Fatal("actionsEqual() = true for different target kinds")
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
