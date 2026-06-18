// Package sim runs many games over the same four decks deterministically. Each
// game derives its own seed from a master seed and builds its own RNG, Engine,
// and Game, so games are independent and reproducible: the same Config always
// produces the same results.
package sim

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"runtime/debug"
	"sync"

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
// games to run, the master seed all per-game seeds derive from, how to build the
// agents for each game, and how many games to run concurrently.
type Config struct {
	Configs   [game.NumPlayers]game.PlayerConfig
	Games     int
	Seed      uint64
	NewAgents AgentFactory

	// Workers is the maximum number of games to run concurrently. Zero or
	// negative means GOMAXPROCS. Because every game is independent and its result
	// is stored at its own index, the returned slice is identical regardless of
	// the worker count. A NewAgents factory must therefore be safe to call from
	// multiple goroutines.
	Workers int
}

// Run plays cfg.Games games over the four configs and returns a SimulationResult
// holding every game's result in order plus the per-game seeds. It is
// deterministic: the same Config yields an identical result regardless of the
// worker count, because each game is independent and its result is written to
// its own index. A nil NewAgents seats a deterministic FirstLegal agent for
// every player.
//
// A game that panics (an engine bug, an unsupported card, or an illegal action)
// is recovered and recorded in Failures attributed to its index and seed; its
// slot in Games holds the zero result and the rest of the batch still completes.
func Run(cfg Config) SimulationResult {
	result := SimulationResult{
		Games:      make([]rules.GameResult, cfg.Games),
		Seeds:      make([]uint64, cfg.Games),
		GameCount:  cfg.Games,
		MasterSeed: cfg.Seed,
	}
	for i := range cfg.Games {
		result.Seeds[i] = GameSeed(cfg.Seed, i)
	}
	games := result.Games
	// failures[i] is the failure of game i, or nil; distinct indices write
	// disjoint slots, so collection stays race-free and order-independent.
	failures := make([]*GameFailure, cfg.Games)

	workers := workerCount(cfg)
	if workers <= 1 {
		for i := range cfg.Games {
			games[i], failures[i] = runGameSafely(cfg, i)
		}
		return withFailures(result, failures)
	}

	jobs := make(chan int)
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for index := range jobs {
				games[index], failures[index] = runGameSafely(cfg, index)
			}
		})
	}
	for i := range cfg.Games {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return withFailures(result, failures)
}

// runGameSafely runs one game and recovers a panic so a single failing game does
// not abort the batch. On a panic it returns the zero result and a GameFailure
// attributing the panic to the game's index and seed.
func runGameSafely(cfg Config, index int) (result rules.GameResult, failure *GameFailure) {
	defer func() {
		if r := recover(); r != nil {
			result = rules.GameResult{}
			failure = &GameFailure{
				Index:  index,
				Seed:   GameSeed(cfg.Seed, index),
				Reason: fmt.Sprintf("%v", r),
				Stack:  string(debug.Stack()),
			}
		}
	}()
	return RunOne(cfg, index), nil
}

// withFailures gathers the per-game failure slots into result.Failures in index
// order, so the failure list is deterministic regardless of completion order.
func withFailures(result SimulationResult, failures []*GameFailure) SimulationResult {
	for i := range failures {
		if failures[i] != nil {
			result.Failures = append(result.Failures, *failures[i])
		}
	}
	return result
}

// workerCount resolves the configured worker count against the batch size: at
// most one worker per game, defaulting to GOMAXPROCS when unset.
func workerCount(cfg Config) int {
	workers := cfg.Workers
	if workers <= 0 {
		workers = runtime.GOMAXPROCS(0)
	}
	if workers > cfg.Games {
		workers = cfg.Games
	}
	return workers
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
