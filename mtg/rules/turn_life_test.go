package rules

import (
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestTurnLogCapturesLifeTotals confirms each turn records every player's life
// at the start of the turn (the Commander starting total of 40 here, since the
// pure-land game deals no damage).
func TestTurnLogCapturesLifeTotals(t *testing.T) {
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
		configs[i] = game.PlayerConfig{Name: "Seat", Commander: commander, Deck: repeatedCard(forest, 99)}
	}

	engine := NewEngine(rand.New(rand.NewPCG(1, 9)))
	g := engine.NewGame(configs)
	result := engine.RunGameWithTurnLimit(g, [game.NumPlayers]PlayerAgent{}, 8)

	if len(result.Turns) == 0 {
		t.Fatal("no turns were played")
	}
	for _, turn := range result.Turns {
		for seat := range turn.LifeTotals {
			if turn.LifeTotals[seat] != 40 {
				t.Fatalf("turn %d seat %d life = %d, want 40", turn.TurnNumber, seat, turn.LifeTotals[seat])
			}
		}
	}
}
