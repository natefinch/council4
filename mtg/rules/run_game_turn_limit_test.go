package rules

import (
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestRunGameWithTurnLimitStopsAtLimit plays a four-player game of pure lands
// (no win condition) and confirms the turn limit caps the game instead of
// running to the 1000-turn safety cap, leaving no winner.
func TestRunGameWithTurnLimitStopsAtLimit(t *testing.T) {
	forest := &game.CardDef{CardFace: game.CardFace{
		Name:       "Forest",
		Supertypes: []types.Super{types.Basic},
		Types:      []types.Card{types.Land},
	}}
	commander := &game.CardDef{CardFace: game.CardFace{
		Name:       "Idle Commander",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
	}}
	var configs [game.NumPlayers]game.PlayerConfig
	for i := range configs {
		configs[i] = game.PlayerConfig{
			Name:      "Seat",
			Commander: commander,
			Deck:      repeatedCard(forest, 99),
		}
	}

	engine := NewEngine(rand.New(rand.NewPCG(3, 5)))
	g := engine.NewGame(configs)
	result := engine.RunGameWithTurnLimit(g, [game.NumPlayers]PlayerAgent{}, 12)

	if result.TurnCount > 12 {
		t.Fatalf("turn count = %d, want at most the limit of 12", result.TurnCount)
	}
	if result.HasWinner {
		t.Fatal("a no-win game should reach the turn limit with no winner")
	}
}
