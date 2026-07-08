package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestCantBecomeMonarchThisTurnBlocksDesignation proves "You can't become the
// monarch this turn." (Jared Carthalion): once the flag is set, setMonarch is a
// no-op for that player (including the combat-damage steal), and the flag clears
// as the next turn begins.
func TestCantBecomeMonarchThisTurnBlocksDesignation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].CantBecomeMonarchThisTurn = true

	if setMonarch(g, game.Player1) {
		t.Fatal("setMonarch succeeded while player can't become monarch, want no-op")
	}
	if g.Players[game.Player1].IsMonarch {
		t.Fatal("player became monarch despite the restriction")
	}

	// Another player is unaffected.
	if !setMonarch(g, game.Player2) {
		t.Fatal("setMonarch(Player2) = false, want true")
	}
	if !g.Players[game.Player2].IsMonarch {
		t.Fatal("Player2 did not become monarch")
	}

	// The restriction clears as the next turn begins.
	NewEngine(nil).advanceToNextTurn(g)
	if g.Players[game.Player1].CantBecomeMonarchThisTurn {
		t.Fatal("restriction not cleared at turn start")
	}
	if !setMonarch(g, game.Player1) {
		t.Fatal("setMonarch(Player1) still blocked after turn advance")
	}
}
