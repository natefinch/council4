package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestDynamicAmountCommanderCastCount verifies that the commander-cast count
// reads the resolving ability controller's CommanderCastCount and scales by the
// amount multiplier, backing the command-zone-cast anthem family (Commander's
// Insignia; Vanguard of the Restless).
func TestDynamicAmountCommanderCastCount(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := &game.StackObject{Controller: game.Player1}

	g.Players[game.Player1].CommanderCastCount = 3
	g.Players[game.Player2].CommanderCastCount = 5
	if got := dynamicAmountValue(g, obj, game.Player1, game.DynamicAmount{
		Kind:       game.DynamicAmountCommanderCastCount,
		Multiplier: 1,
	}); got != 3 {
		t.Fatalf("commander cast count = %d, want 3", got)
	}

	g.Players[game.Player1].CommanderCastCount = 0
	if got := dynamicAmountValue(g, obj, game.Player1, game.DynamicAmount{
		Kind:       game.DynamicAmountCommanderCastCount,
		Multiplier: 1,
	}); got != 0 {
		t.Fatalf("zero commander cast count = %d, want 0", got)
	}
}
