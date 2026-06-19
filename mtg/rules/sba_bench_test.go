package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// BenchmarkStableStateBasedActionsLargeBoard measures the common no-mutation
// pass over a late-game board. Every priority point runs this pass, so derived
// permanent values must be reused across its attachment and legendary scans.
func BenchmarkStableStateBasedActionsLargeBoard(b *testing.B) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	card := &game.CardDef{CardFace: game.CardFace{
		Name:  "Benchmark Permanent",
		Types: []types.Card{types.Land},
	}}
	for range 99 {
		addPermanentForSBA(g, game.Player1, card)
	}

	b.ResetTimer()
	for range b.N {
		engine.applyStateBasedActions(g)
	}
}
