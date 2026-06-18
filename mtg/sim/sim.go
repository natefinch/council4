// Package sim runs many games over the same four decks deterministically. Each
// game derives its own seed from a master seed and builds its own RNG, Engine,
// and Game, so games are independent and reproducible: the same Config always
// produces the same results.
package sim

import (
	"math/rand/v2"

	"github.com/natefinch/council4/mtg/agent"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/rules"
)

// seedGamma is the odd golden-ratio increment used to spread successive game
// seeds across the 64-bit space before mixing (the SplitMix64 construction).
const seedGamma = 0x9e3779b97f4a7c15

// AgentFactory builds the four seated agents for one game. It receives that
// game's derived seed so any agent randomness stays reproducible; give each seat
// its own RNG derived from gameSeed rather than sharing one.
type AgentFactory func(gameSeed uint64) [game.NumPlayers]rules.PlayerAgent

// Config describes a simulation run: the four player configs to play, how many
// games to run, the master seed all per-game seeds derive from, and how to build
// the agents for each game.
type Config struct {
	Configs   [game.NumPlayers]game.PlayerConfig
	Games     int
	Seed      uint64
	NewAgents AgentFactory
}

// Run plays cfg.Games games over the four configs and returns every game's
// result in order. It is deterministic: the same Config yields identical
// results. A nil NewAgents seats a deterministic FirstLegal agent for every
// player.
func Run(cfg Config) []rules.GameResult {
	results := make([]rules.GameResult, cfg.Games)
	for i := range cfg.Games {
		results[i] = RunOne(cfg, i)
	}
	return results
}

// RunOne plays a single game by index and returns its result. The index selects
// the per-game seed, so RunOne(cfg, i) always reproduces the same game.
func RunOne(cfg Config, index int) rules.GameResult {
	gameSeed := GameSeed(cfg.Seed, index)
	engine := rules.NewEngine(NewRand(gameSeed))
	g := engine.NewGame(cfg.Configs)
	agents := newAgents(cfg)(gameSeed)
	return *engine.RunGame(g, agents)
}

// GameSeed derives the seed for the index-th game from a master seed. Successive
// indices are spread by the golden-ratio increment and mixed with SplitMix64, so
// neighbouring games get well-separated, uncorrelated seeds. A negative index is
// treated as zero.
func GameSeed(master uint64, index int) uint64 {
	offset := uint64(0)
	if index > 0 {
		offset = uint64(index)
	}
	return splitMix64(master + offset*seedGamma)
}

// NewRand builds the RNG for one game from its seed, giving each game an
// independent stream so games never share mutable RNG state.
func NewRand(seed uint64) *rand.Rand {
	return rand.New(rand.NewPCG(seed, seed^seedGamma))
}

func newAgents(cfg Config) AgentFactory {
	if cfg.NewAgents != nil {
		return cfg.NewAgents
	}
	return func(uint64) [game.NumPlayers]rules.PlayerAgent {
		var agents [game.NumPlayers]rules.PlayerAgent
		for i := range agents {
			agents[i] = agent.FirstLegal{}
		}
		return agents
	}
}

// splitMix64 is the SplitMix64 finalizer (Steele et al.), a fast bijective mix
// that turns a counter into a well-distributed 64-bit seed.
func splitMix64(x uint64) uint64 {
	x += seedGamma
	x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
	x = (x ^ (x >> 27)) * 0x94d049bb133111eb
	return x ^ (x >> 31)
}
