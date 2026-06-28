package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// TestGiveControlToTargetPlayerTransfersControl proves the Donate form: a spell
// controlled by Player1 makes a target player (Player2) gain control of a
// permanent Player1 controls. The control-layer continuous effect resolves its
// NewControllerRef to the chosen target player, so the permanent's effective
// controller becomes Player2 — never the resolving controller, Player1.
func TestGiveControlToTargetPlayerTransfersControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCreaturePermanent(g, game.Player1)
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Primitive: game.ApplyContinuous{
			Object: opt.Val(game.TargetPermanentReference(1)),
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:            game.LayerControl,
				NewControllerRef: opt.Val(game.TargetPlayerReference(0)),
			}},
			Duration: game.DurationPermanent,
		},
	}}, []game.Target{
		game.PlayerTarget(game.Player2),
		game.PermanentTarget(creature.ObjectID),
	})

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := effectiveController(g, creature); got != game.Player2 {
		t.Fatalf("effective controller = %v, want Player2 (the chosen target player)", got)
	}
}

// TestGiveControlOfSourceToTargetPlayer proves the Jinxed Idol form: the
// resolving source permanent itself is handed to a target player. The
// continuous effect addresses the source object and resolves its new controller
// to the chosen target player.
func TestGiveControlOfSourceToTargetPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCreaturePermanent(g, game.Player1)
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackActivatedAbility,
		SourceID:   source.ObjectID,
		Controller: game.Player1,
		Targets:    []game.Target{game.PlayerTarget(game.Player2)},
	}
	applyTypedContinuousEffects(g, obj, source, []game.ContinuousEffect{{
		Layer:            game.LayerControl,
		NewControllerRef: opt.Val(game.TargetPlayerReference(0)),
	}}, game.DurationPermanent)

	if got := effectiveController(g, source); got != game.Player2 {
		t.Fatalf("effective controller = %v, want Player2 (the chosen target player)", got)
	}
}
