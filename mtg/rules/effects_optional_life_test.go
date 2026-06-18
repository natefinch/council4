package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

// optionalMayAgent answers the optional "Apply optional effect?" may-choice
// according to accept and passes on every action.
type optionalMayAgent struct {
	accept bool
}

func (optionalMayAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a optionalMayAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceMay && a.accept {
		return []int{1}
	}
	return []int{0}
}

// TestOptionalGainLifeDeclineSkips verifies that declining an Optional GainLife
// instruction leaves the controller's life total unchanged.
func TestOptionalGainLifeDeclineSkips(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	before := g.Players[game.Player1].Life
	addInstructionSpellToStack(g, []game.Instruction{{
		Optional:  true,
		Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(3)},
	}})
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: false}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != before {
		t.Fatalf("life = %d, want %d (declining must skip the gain)", got, before)
	}
}

// TestOptionalGainLifeAcceptPerforms verifies that accepting an Optional
// GainLife instruction adds the life to the controller's total.
func TestOptionalGainLifeAcceptPerforms(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	before := g.Players[game.Player1].Life
	addInstructionSpellToStack(g, []game.Instruction{{
		Optional:  true,
		Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(3)},
	}})
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != before+3 {
		t.Fatalf("life = %d, want %d (accepting must gain 3)", got, before+3)
	}
}
