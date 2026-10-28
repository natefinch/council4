package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// benchBoardGame builds a battlefield of n vanilla lands plus a few creatures,
// approximating the heavy board a long Commander game accumulates. It returns
// the game so a benchmark can repeatedly query effective values.
func benchBoardGame(n int) *game.Game {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	land := &game.CardDef{CardFace: game.CardFace{
		Name:     "Forest",
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Forest},
	}}
	creature := &game.CardDef{CardFace: game.CardFace{
		Name:      "Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}}
	for i := range n {
		def := land
		if i%4 == 0 {
			def = creature
		}
		addCombatPermanent(g, game.PlayerID(i%game.NumPlayers), def)
	}
	return g
}

// benchPriorityPointRecompute approximates the work done at a single priority
// point: every permanent's effective characteristics are queried once. This is
// the operation whose cost grows with the board and dominates long games.
func benchPriorityPointRecompute(g *game.Game) int {
	total := 0
	for _, permanent := range g.Battlefield {
		total += effectivePower(g, permanent)
		if _, ok := effectiveToughness(g, permanent); ok {
			total++
		}
	}
	return total
}

func BenchmarkEffectiveValuesBoard50(b *testing.B) {
	benchmarkEffectiveValuesBoard(b, 50)
}

func BenchmarkEffectiveValuesBoard99(b *testing.B) {
	benchmarkEffectiveValuesBoard(b, 99)
}

func BenchmarkEffectiveValuesBoard200(b *testing.B) {
	benchmarkEffectiveValuesBoard(b, 200)
}

func benchmarkEffectiveValuesBoard(b *testing.B, n int) {
	b.Helper()
	g := benchBoardGame(n)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = benchPriorityPointRecompute(g)
	}
}

// BenchmarkEffectiveValuesBoardFramed99 measures the same priority-point
// recompute inside a static-source frame, the way the engine evaluates a
// priority point. The frame builds the static-ability source set once instead
// of rescanning the battlefield for every permanent.
func BenchmarkEffectiveValuesBoardFramed99(b *testing.B) {
	benchmarkEffectiveValuesBoardFramed(b, 99)
}

func BenchmarkEffectiveValuesBoardFramed200(b *testing.B) {
	benchmarkEffectiveValuesBoardFramed(b, 200)
}

func benchmarkEffectiveValuesBoardFramed(b *testing.B, n int) {
	b.Helper()
	g := benchBoardGame(n)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		g.BeginStaticSourceFrame()
		_ = benchPriorityPointRecompute(g)
		g.EndStaticSourceFrame()
	}
}

func BenchmarkEffectiveValuesJetmir(b *testing.B) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := jetmirPermanent(g, game.Player1)
	for range 8 {
		addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	}
	if got := effectivePower(g, source); got != 6 {
		b.Fatalf("Jetmir power = %d, want all three thresholds active", got)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = effectivePermanentValues(g, source)
	}
}

func BenchmarkEffectiveValuesCopiedOpalescence(b *testing.B) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	opal := addCombatPermanent(g, game.Player1, authored.Opalescence.New())
	blueprint := addCombatPermanent(g, game.Player1, challengeEnchantment("Five mana blueprint", 5))
	copied := addCombatPermanent(g, game.Player2, challengeEnchantment("Copy holder", 0))
	becomeStaticCopy(b, g, copied, blueprint, false)
	becomeStaticCopy(b, g, copied, opal, false)
	addCombatPermanent(g, game.Player1, authored.Humility.New())
	if got := effectivePower(g, copied); got != 1 {
		b.Fatalf("Humility copy power = %d, want 1", got)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = effectivePermanentValues(g, copied)
	}
}
