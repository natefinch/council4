package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestGroupSourceDamageHitsEachController verifies that GroupSourceDamage has
// every creature deal its amount to the player who controls it ("Each creature
// deals 1 damage to its controller."): each controller loses life once per
// creature they control.
func TestGroupSourceDamageHitsEachController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	startP1 := g.Players[game.Player1].Life
	startP2 := g.Players[game.Player2].Life
	addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	addCombatCreaturePermanentWithPower(g, game.Player2, 3)

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Primitive: game.GroupSourceDamage{
			Group: game.BattlefieldGroup(game.Selection{
				RequiredTypes: []types.Card{types.Creature},
			}),
			Amount: game.Fixed(1),
		},
	}}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != startP1-1 {
		t.Fatalf("Player1 life = %d, want %d (1 creature)", got, startP1-1)
	}
	if got := g.Players[game.Player2].Life; got != startP2-2 {
		t.Fatalf("Player2 life = %d, want %d (2 creatures)", got, startP2-2)
	}
}
