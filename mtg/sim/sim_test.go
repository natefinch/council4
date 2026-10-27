package sim

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/natefinch/council4/mtg/agent"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
)

func forest() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Forest",
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Forest},
	}}
}

// smokeConfigs builds four tiny land-only decks that play out quickly.
func smokeConfigs() [game.NumPlayers]game.PlayerConfig {
	var configs [game.NumPlayers]game.PlayerConfig
	for player := range configs {
		for range 12 {
			configs[player].Deck = append(configs[player].Deck, forest())
		}
	}
	return configs
}

func smokeConfig(games int, seed uint64) Config {
	configs := smokeConfigs()
	for player := range configs {
		// One draw beyond the opening hand, then a real deck-out result.
		configs[player].Deck = configs[player].Deck[:8]
	}
	return Config{
		Configs:   configs,
		Games:     games,
		Seed:      seed,
		NewAgents: nil, // default FirstLegal agents
	}
}

func TestRunIsDeterministicForAMasterSeed(t *testing.T) {
	cfg := smokeConfig(3, 12345)
	first := Run(cfg)
	second := Run(cfg)
	requireCompletedGames(t, first)
	requireCompletedGames(t, second)

	if len(first.Games) != cfg.Games || len(second.Games) != cfg.Games {
		t.Fatalf("Run returned %d/%d results, want %d", len(first.Games), len(second.Games), cfg.Games)
	}
	if !reflect.DeepEqual(first, second) {
		t.Error("Run is not deterministic: two runs with the same master seed differ")
	}
}

func TestRunOneReproducesTheSameGame(t *testing.T) {
	cfg := smokeConfig(3, 999)
	all := Run(cfg)
	requireCompletedGames(t, all)
	for i := range cfg.Games {
		if again := RunOne(cfg, i); !reflect.DeepEqual(again, all.Games[i]) {
			t.Errorf("RunOne(cfg, %d) does not reproduce game %d from Run", i, i)
		}
	}
}

func TestDifferentMasterSeedsCanDiffer(t *testing.T) {
	// Not a strict guarantee for every pair, but a non-degenerate seed mixer
	// should make at least one game of a small batch differ across seeds.
	a := Run(smokeConfig(2, 1))
	b := Run(smokeConfig(2, 2))
	requireCompletedGames(t, a)
	requireCompletedGames(t, b)
	if reflect.DeepEqual(a.Games, b.Games) {
		t.Error("two different master seeds produced identical batches; seed derivation looks degenerate")
	}
}

func TestGameSeedDerivationIsDistinctAndStable(t *testing.T) {
	const master = 42
	seen := make(map[uint64]int)
	for i := range 256 {
		seed := GameSeed(master, i)
		if prev, dup := seen[seed]; dup {
			t.Fatalf("GameSeed collision: index %d and %d both produced %d", prev, i, seed)
		}
		seen[seed] = i
		if GameSeed(master, i) != seed {
			t.Fatalf("GameSeed(%d, %d) is not stable", master, i)
		}
	}
}

func TestRunHonorsAgentFactory(t *testing.T) {
	// A custom factory is invoked once per game with that game's derived seed.
	// Force a single worker so the recorded order is the game order.
	var seeds []uint64
	cfg := smokeConfig(3, 7)
	cfg.Workers = 1
	cfg.NewAgents = func(gameSeed uint64) [game.NumPlayers]rules.PlayerAgent {
		seeds = append(seeds, gameSeed)
		var agents [game.NumPlayers]rules.PlayerAgent
		for i := range agents {
			agents[i] = agent.FirstLegal{}
		}
		return agents
	}
	Run(cfg)

	if len(seeds) != cfg.Games {
		t.Fatalf("agent factory called %d times, want %d", len(seeds), cfg.Games)
	}
	for i := range cfg.Games {
		if seeds[i] != GameSeed(cfg.Seed, i) {
			t.Errorf("game %d agent seed = %d, want %d", i, seeds[i], GameSeed(cfg.Seed, i))
		}
	}
}

// TestParallelMatchesSequential checks that the aggregate is identical across
// worker counts for a fixed master seed. Run under -race, the parallel path also
// proves games share no mutable state.
func TestParallelMatchesSequential(t *testing.T) {
	base := smokeConfig(3, 4242)

	sequential := base
	sequential.Workers = 1
	want := Run(sequential)
	requireCompletedGames(t, want)

	for _, workers := range []int{2, 8} {
		cfg := base
		cfg.Workers = workers
		got := Run(cfg)
		requireCompletedGames(t, got)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Run with %d workers differs from sequential", workers)
		}
	}
}

// TestDefaultWorkersRunsConcurrently checks the default (zero) worker count runs
// a batch correctly; combined with -race it exercises the concurrent path.
func TestDefaultWorkersRunsConcurrently(t *testing.T) {
	cfg := smokeConfig(3, 55) // Workers left at 0 -> GOMAXPROCS
	got := Run(cfg)
	requireCompletedGames(t, got)
	if len(got.Games) != cfg.Games {
		t.Fatalf("Run returned %d results, want %d", len(got.Games), cfg.Games)
	}
	sequential := cfg
	sequential.Workers = 1
	if want := Run(sequential); !reflect.DeepEqual(got, want) {
		t.Error("default-worker run differs from the sequential run")
	}
}

func TestWorkerCount(t *testing.T) {
	for _, test := range []struct {
		name           string
		games, workers int
		want           int
	}{
		{"default", 3, 0, min(3, runtime.GOMAXPROCS(0))},
		{"negative", 3, -1, min(3, runtime.GOMAXPROCS(0))},
		{"sequential", 3, 1, 1},
		{"parallel", 3, 2, 2},
		{"clamped", 3, 32, 3},
		{"empty", 0, 4, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := workerCount(Config{Games: test.games, Workers: test.workers}); got != test.want {
				t.Fatalf("workerCount = %d, want %d", got, test.want)
			}
		})
	}
}

func requireCompletedGames(t *testing.T, result SimulationResult) {
	t.Helper()
	if result.FailureCount() != 0 {
		t.Fatalf("fixture games failed: %+v", result.Failures)
	}
	for i, game := range result.Games {
		if !game.Completed() || !game.HasWinner || game.TurnCount == 0 {
			t.Fatalf("fixture game %d did not reach a rules win: %s after %d turns", i, game.Outcome(), game.TurnCount)
		}
	}
}
