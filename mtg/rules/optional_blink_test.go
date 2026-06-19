package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// optionalBlinkInstructions builds the exile-then-return instruction pair the
// cardgen optional-blink lowerer emits for "you may exile target creature you
// control, then return that card to the battlefield under its owner's control":
// the exile is Optional and publishes its result under "if-you-do", and the
// return is gated on that exile having succeeded.
func optionalBlinkInstructions() []game.Instruction {
	return []game.Instruction{
		{
			Primitive:     game.Exile{Object: game.TargetPermanentReference(0), ExileLinkedKey: "blink"},
			Optional:      true,
			PublishResult: game.ResultKey("if-you-do"),
		},
		{
			Primitive: game.PutOnBattlefield{Source: game.LinkedBattlefieldSource("blink")},
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:       game.ResultKey("if-you-do"),
				Succeeded: game.TriTrue,
			}),
		},
	}
}

func resolveOptionalBlink(t *testing.T, accept bool) (*game.Game, *game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Blinked Creature",
		Types: []types.Card{types.Creature},
	}})
	addInstructionSpellToStackForController(g, game.Player1, optionalBlinkInstructions(),
		[]game.Target{game.PermanentTarget(target.ObjectID)})
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: accept}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	return g, target
}

// TestOptionalBlinkDeclineSkips verifies that declining the optional exile leaves
// the targeted creature untouched on the battlefield: the gated return never
// runs, so no card leaves and re-enters. This proves the decline branch of the
// controller's choice.
func TestOptionalBlinkDeclineSkips(t *testing.T) {
	g, target := resolveOptionalBlink(t, false)
	remaining, ok := permanentByObjectID(g, target.ObjectID)
	if !ok {
		t.Fatal("declining the optional exile must leave the original creature on the battlefield")
	}
	if remaining.ObjectID != target.ObjectID {
		t.Fatalf("creature object identity changed to %v, want unchanged %v", remaining.ObjectID, target.ObjectID)
	}
	if len(g.Battlefield) != 1 {
		t.Fatalf("battlefield = %d permanents, want exactly the untouched creature", len(g.Battlefield))
	}
}

// TestOptionalBlinkAcceptPerforms verifies that accepting the optional exile
// exiles the targeted creature and the gated return brings it back as a new
// object. This proves the accept branch of the controller's choice.
func TestOptionalBlinkAcceptPerforms(t *testing.T) {
	g, target := resolveOptionalBlink(t, true)
	if _, ok := permanentByObjectID(g, target.ObjectID); ok {
		t.Fatal("accepting the optional exile must remove the original creature object")
	}
	var returned *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == target.CardInstanceID {
			returned = permanent
		}
	}
	if returned == nil {
		t.Fatal("accepting the optional exile must return the blinked creature to the battlefield")
	}
	if returned.ObjectID == target.ObjectID {
		t.Fatal("returned permanent reused the original object identity, want a new object")
	}
	if returned.Controller != game.Player1 || returned.Owner != game.Player1 {
		t.Fatalf("returned controller/owner = %v/%v, want Player1/Player1", returned.Controller, returned.Owner)
	}
}
