package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestOptionalDamageDeclineSkips verifies that declining an Optional Damage
// instruction (the "you may have it deal N damage to <target>" causative)
// leaves the targeted player's life total unchanged.
func TestOptionalDamageDeclineSkips(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	before := g.Players[game.Player2].Life
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Optional:  true,
		Primitive: game.Damage{Amount: game.Fixed(2), Recipient: game.AnyTargetDamageRecipient(0)},
	}}, []game.Target{game.PlayerTarget(game.Player2)})
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: false}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != before {
		t.Fatalf("life = %d, want %d (declining must skip the damage)", got, before)
	}
}

// TestOptionalDamageAcceptPerforms verifies that accepting an Optional Damage
// instruction deals the damage to the targeted player.
func TestOptionalDamageAcceptPerforms(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	before := g.Players[game.Player2].Life
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Optional:  true,
		Primitive: game.Damage{Amount: game.Fixed(2), Recipient: game.AnyTargetDamageRecipient(0)},
	}}, []game.Target{game.PlayerTarget(game.Player2)})
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != before-2 {
		t.Fatalf("life = %d, want %d (accepting must deal 2 damage)", got, before-2)
	}
}
