package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestExileTopOfLibraryControllerExilesTopCard proves a controller-scoped
// "exile the top card of your library" moves the top library card to the
// controller's exile zone.
func TestExileTopOfLibraryControllerExilesTopCard(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.ExileTopOfLibrary{
		Amount: game.Fixed(1),
		Player: game.ControllerReference(),
	}, nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Next"}})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Library.Size(); got != 1 {
		t.Fatalf("library size = %d, want 1", got)
	}
	if got := g.Players[game.Player1].Exile.Size(); got != 1 {
		t.Fatalf("exile size = %d, want 1", got)
	}
}

// TestExileTopOfLibraryGroupExilesForEveryPlayer proves "each player exiles the
// top two cards" moves cards from every player's library to their exile zone.
func TestExileTopOfLibraryGroupExilesForEveryPlayer(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.ExileTopOfLibrary{
		Amount:      game.Fixed(2),
		PlayerGroup: game.AllPlayersReference(),
	}, nil)
	for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		addCardToLibrary(g, playerID, &game.CardDef{CardFace: game.CardFace{Name: "A"}})
		addCardToLibrary(g, playerID, &game.CardDef{CardFace: game.CardFace{Name: "B"}})
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		if got := g.Players[playerID].Library.Size(); got != 0 {
			t.Fatalf("player %d library size = %d, want 0", playerID, got)
		}
		if got := g.Players[playerID].Exile.Size(); got != 2 {
			t.Fatalf("player %d exile size = %d, want 2", playerID, got)
		}
	}
}
