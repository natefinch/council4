package sim

import (
	"reflect"
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
	return Config{
		Configs:   smokeConfigs(),
		Games:     games,
		Seed:      seed,
		NewAgents: nil, // default FirstLegal agents
	}
}

func TestRunIsDeterministicForAMasterSeed(t *testing.T) {
	cfg := smokeConfig(8, 12345)
	first := Run(cfg)
	second := Run(cfg)

	if len(first) != cfg.Games || len(second) != cfg.Games {
		t.Fatalf("Run returned %d/%d results, want %d", len(first), len(second), cfg.Games)
	}
	if !reflect.DeepEqual(first, second) {
		t.Error("Run is not deterministic: two runs with the same master seed differ")
	}
}

func TestRunOneReproducesTheSameGame(t *testing.T) {
	cfg := smokeConfig(4, 999)
	all := Run(cfg)
	for i := range cfg.Games {
		if again := RunOne(cfg, i); !reflect.DeepEqual(again, all[i]) {
			t.Errorf("RunOne(cfg, %d) does not reproduce game %d from Run", i, i)
		}
	}
}

func TestDifferentMasterSeedsCanDiffer(t *testing.T) {
	// Not a strict guarantee for every pair, but a non-degenerate seed mixer
	// should make at least one game of a small batch differ across seeds.
	a := Run(smokeConfig(8, 1))
	b := Run(smokeConfig(8, 2))
	if reflect.DeepEqual(a, b) {
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
	var seeds []uint64
	cfg := smokeConfig(3, 7)
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
