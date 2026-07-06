package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestMillTargetedPlayersMillsEachChosenTarget proves that a group mill naming
// the TargetedPlayers group ("any number of target players each mill two cards"
// — Court of Cunning) mills every player chosen as a target and leaves untargeted
// players untouched.
func TestMillTargetedPlayersMillsEachChosenTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		for range 5 {
			addCardToLibraryNamed(g, playerID, "Lib")
		}
	}

	addEffectSpellToStack(g, game.Player1, game.Mill{
		Amount:      game.Fixed(2),
		PlayerGroup: game.TargetedPlayersReference(),
	}, []game.Target{
		{Kind: game.TargetPlayer, PlayerID: game.Player2},
		{Kind: game.TargetPlayer, PlayerID: game.Player3},
	})

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := g.Players[game.Player2].Library.Size(); got != 3 {
		t.Errorf("Player2 library = %d, want 3 (milled 2)", got)
	}
	if got := g.Players[game.Player3].Library.Size(); got != 3 {
		t.Errorf("Player3 library = %d, want 3 (milled 2)", got)
	}
	if got := g.Players[game.Player1].Library.Size(); got != 5 {
		t.Errorf("Player1 library = %d, want 5 (not targeted)", got)
	}
	if got := g.Players[game.Player4].Library.Size(); got != 5 {
		t.Errorf("Player4 library = %d, want 5 (not targeted)", got)
	}
}
