package sim

import (
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/agent"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/rules"
)

// randomAgentsFactory seats a RandomAgent per player, each with its own RNG
// derived from the game seed. Random agents exercise RNG-dependent branches the
// deterministic FirstLegal agent never reaches, so determinism here is a
// stronger guarantee. Building fresh per-seat RNGs on every call keeps the
// factory safe to invoke from multiple parallel workers.
func randomAgentsFactory(gameSeed uint64) [game.NumPlayers]rules.PlayerAgent {
	var agents [game.NumPlayers]rules.PlayerAgent
	for seat := range agents {
		rng := rand.New(rand.NewPCG(gameSeed, uint64(seat)+1))
		agents[seat] = agent.NewRandomAgent(rng)
	}
	return agents
}

func randomAgentConfig(games int, seed uint64) Config {
	return Config{
		Configs:   smokeConfigs(),
		Games:     games,
		Seed:      seed,
		NewAgents: randomAgentsFactory,
	}
}

// TestRunWithRandomAgentsIsDeterministic strengthens the same-seed guarantee to
// RNG-driven agents: two runs of the same master seed must be identical even
// when every decision is drawn from a random source.
func TestRunWithRandomAgentsIsDeterministic(t *testing.T) {
	cfg := randomAgentConfig(12, 13579)

	first := Run(cfg)
	second := Run(cfg)

	if !reflect.DeepEqual(first, second) {
		t.Error("Run with random agents is not deterministic across two runs of the same master seed")
	}
}

// TestRunParallelMatchesSequential proves the worker count never affects the
// result: a fixed master seed run on a single worker and on many workers must
// produce byte-identical SimulationResults. This also runs under -race in CI,
// guarding the concurrent collection path.
func TestRunParallelMatchesSequential(t *testing.T) {
	sequential := randomAgentConfig(16, 7777)
	sequential.Workers = 1
	parallel := randomAgentConfig(16, 7777)
	parallel.Workers = 8

	seq := Run(sequential)
	par := Run(parallel)

	if !reflect.DeepEqual(seq, par) {
		t.Error("parallel and sequential simulation produced different results for the same master seed")
	}
}

// TestRunOneIsIdenticalAcrossRepeats guards against map-iteration order (which
// Go randomizes on every range, even within one process) leaking into a
// slice-valued GameResult field. Replaying the same game many times in one
// process must yield an identical result every time; if any result slice were
// built by ranging a map, these repeats would diverge.
func TestRunOneIsIdenticalAcrossRepeats(t *testing.T) {
	cfg := randomAgentConfig(1, 24680)

	first := RunOne(cfg, 0)
	for repeat := range 30 {
		again := RunOne(cfg, 0)
		if !reflect.DeepEqual(again, first) {
			t.Fatalf("RunOne repeat %d differs from the first run; nondeterminism (e.g. map-iteration order) is leaking into GameResult", repeat)
		}
	}
}
