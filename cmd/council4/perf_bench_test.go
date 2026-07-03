package main

import (
	"path/filepath"
	"testing"

	"github.com/natefinch/council4/mtg/agent"
	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/deck"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/mtg/sim"
)

// perfDeckConfigs loads the committed realistic baseline decks under
// testdata/perf and resolves them against the full default registry. These
// decks are legal four-player Commander decks built from supported cards (see
// docs/perf/README.md); they give a repeatable game-time baseline for engine
// performance work.
func perfDeckConfigs(tb testing.TB) [game.NumPlayers]game.PlayerConfig {
	tb.Helper()
	var inputs [game.NumPlayers]deck.PlayerInput
	for i := range inputs {
		path := filepath.Join("testdata", "perf", "deck"+string(rune('1'+i))+".txt")
		decklist, err := deck.ParseFile(path)
		if err != nil {
			tb.Fatalf("parse %s: %v", path, err)
		}
		inputs[i] = deck.PlayerInput{Name: path, Decklist: decklist}
	}
	loaded := deck.Load(inputs, game.Player1, cards.NewDefaultRegistry())
	if !loaded.OK() {
		tb.Fatalf("baseline decks did not load cleanly: %+v unresolved=%+v",
			loaded.Legality, loaded.Unresolved)
	}
	return loaded.Configs
}

func benchmarkPerfDeckGame(b *testing.B, factory sim.AgentFactory) {
	configs := perfDeckConfigs(b)
	cfg := sim.Config{Configs: configs, Games: 1, Seed: 20260619, NewAgents: factory}
	b.ResetTimer()
	for range b.N {
		_ = sim.RunOne(cfg, 0)
	}
}

// BenchmarkPerfDeckGameFirstLegal plays one full game over the baseline decks
// with the deterministic FirstLegal agent.
func BenchmarkPerfDeckGameFirstLegal(b *testing.B) {
	benchmarkPerfDeckGame(b, func(uint64) [game.NumPlayers]rules.PlayerAgent {
		return agents(agent.FirstLegal{})
	})
}

// BenchmarkPerfDeckGameGeneric plays one full game over the baseline decks with
// the rule-based generic Commander strategy, the realistic playtest agent.
func BenchmarkPerfDeckGameGeneric(b *testing.B) {
	benchmarkPerfDeckGame(b, func(uint64) [game.NumPlayers]rules.PlayerAgent {
		return agents(agent.Agent{Strategy: agent.GenericStrategy{}})
	})
}

// BenchmarkPerfDeckGameSearch plays one full game over the baseline decks with
// the search agent (one-ply lookahead + position evaluation) in every seat. It
// is the search agent's game-time baseline; search is compute-heavy, so this is
// expected to be far slower than the heuristic agents.
func BenchmarkPerfDeckGameSearch(b *testing.B) {
	benchmarkPerfDeckGame(b, func(uint64) [game.NumPlayers]rules.PlayerAgent {
		return agents(agent.Searcher{Rollout: agent.GenericStrategy{}})
	})
}
