package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/opt"
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

// TestExileTopOfLibraryGroupPlacesNamedCounter proves the Counter field places
// one named marker counter on each exiled card across every player's library,
// recorded in Game.ExileCounters (Evelyn, the Covetous collection counters).
func TestExileTopOfLibraryGroupPlacesNamedCounter(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.ExileTopOfLibrary{
		Amount:      game.Fixed(1),
		PlayerGroup: game.AllPlayersReference(),
		Counter:     opt.Val(counter.Collection),
	}, nil)
	topIDs := make(map[game.PlayerID]id.ID)
	for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		addCardToLibrary(g, playerID, &game.CardDef{CardFace: game.CardFace{Name: "Bottom"}})
		topIDs[playerID] = addCardToLibrary(g, playerID, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	for playerID, cardID := range topIDs {
		if !g.Players[playerID].Exile.Contains(cardID) {
			t.Fatalf("player %d top card not exiled", playerID)
		}
		if !g.HasExileCounter(cardID, counter.Collection) {
			t.Fatalf("player %d exiled card missing collection counter", playerID)
		}
	}
}
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
