package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"os"

	"github.com/natefinch/council4/mtg/agent"
	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/deck"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/mtg/sim"
)

// seatStreams gives each seat a distinct, fixed RNG stream offset so random
// agents derive independent streams from one per-game seed without an
// int-to-uint64 conversion.
var seatStreams = [game.NumPlayers]uint64{
	0x1d8e4e27c47d124f, 0x9e3779b97f4a7c15, 0x2545f4914f6cdd1d, 0xa0761d6478bd642f,
}

// agentProfiles lists the supported -agent values for help text and validation.
var agentProfiles = []string{"firstlegal", "random", "generic"}

// loadConfigs parses the four decklist files, resolves them against the registry,
// and returns Commander-legal player configs, reporting parse or legality
// problems as an error instead of panicking.
func loadConfigs(paths []string, tested int, registry *cards.Registry) ([game.NumPlayers]game.PlayerConfig, error) {
	var zero [game.NumPlayers]game.PlayerConfig
	if len(paths) != game.NumPlayers {
		return zero, fmt.Errorf("need exactly %d -deck paths, got %d", game.NumPlayers, len(paths))
	}
	if tested < 1 || tested > game.NumPlayers {
		return zero, fmt.Errorf("-tested must be between 1 and %d", game.NumPlayers)
	}
	var inputs [game.NumPlayers]deck.PlayerInput
	for i, path := range paths {
		decklist, err := deck.ParseFile(path)
		if err != nil {
			return zero, fmt.Errorf("decklist %q: %w", path, err)
		}
		inputs[i] = deck.PlayerInput{Name: deckName(path), Decklist: decklist}
	}
	loaded := deck.Load(inputs, game.PlayerID(tested-1), registry)
	if !loaded.OK() {
		return zero, loadProblems(loaded)
	}
	return loaded.Configs, nil
}

// agentFactory returns the per-game agent factory for an -agent profile.
func agentFactory(profile string) (sim.AgentFactory, error) {
	switch profile {
	case "", "firstlegal":
		return func(uint64) [game.NumPlayers]rules.PlayerAgent { return agents(agent.FirstLegal{}) }, nil
	case "generic":
		return func(uint64) [game.NumPlayers]rules.PlayerAgent {
			return agents(agent.Agent{Strategy: agent.GenericStrategy{}})
		}, nil
	case "random":
		return func(gameSeed uint64) [game.NumPlayers]rules.PlayerAgent {
			var seated [game.NumPlayers]rules.PlayerAgent
			for i := range seated {
				seated[i] = agent.NewRandomAgent(rand.New(rand.NewPCG(gameSeed, gameSeed^seatStreams[i])))
			}
			return seated
		}, nil
	default:
		return nil, fmt.Errorf("-agent must be one of %v", agentProfiles)
	}
}

// runDeckSimulation loads the four decklists and plays a multi-game simulation,
// writing a summary to w and, when outPath is set, a JSON report.
func runDeckSimulation(w io.Writer, paths []string, tested, games, workers int, seed uint64, profile, outPath string, registry *cards.Registry) error {
	if games < 1 {
		return fmt.Errorf("-games must be at least 1, got %d", games)
	}
	configs, err := loadConfigs(paths, tested, registry)
	if err != nil {
		return err
	}
	factory, err := agentFactory(profile)
	if err != nil {
		return err
	}
	result := sim.Run(sim.Config{
		Configs:   configs,
		Games:     games,
		Seed:      seed,
		Workers:   workers,
		NewAgents: factory,
	})
	return reportSimulation(w, result, paths, tested, profile, outPath)
}

// reportSimulation writes the text summary and, when outPath is set, the JSON
// report for a finished simulation.
func reportSimulation(w io.Writer, result sim.SimulationResult, paths []string, tested int, profile, outPath string) error {
	printSimSummary(w, result, paths, tested, profile)
	if outPath == "" {
		return nil
	}
	if err := writeSimReport(outPath, result, paths, tested, profile); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(w, "\nWrote JSON report to %s\n", outPath)
	return nil
}

func printSimSummary(w io.Writer, result sim.SimulationResult, paths []string, tested int, profile string) {
	wins := result.WinCounts()
	completed := result.GameCount - result.FailureCount()
	low, high, avg := gameLengthStats(result)

	_, _ = fmt.Fprintf(w, "Simulation: %d games, agent %q, master seed %d\n",
		result.GameCount, profileName(profile), result.MasterSeed)
	_, _ = fmt.Fprintf(w, "Deck under test: %s (seat %d)\n", deckName(paths[tested-1]), tested)
	_, _ = fmt.Fprintln(w, "\nResults by seat:")
	for seat := range wins {
		marker := "  "
		if seat == tested-1 {
			marker = "* "
		}
		_, _ = fmt.Fprintf(w, "%s%s: %d wins (%.1f%%)\n",
			marker, deckName(paths[seat]), wins[seat], percent(wins[seat], result.GameCount))
	}
	_, _ = fmt.Fprintf(w, "\nDraws: %d   Failures: %d\n", result.DrawCount(), result.FailureCount())
	if completed > 0 {
		_, _ = fmt.Fprintf(w, "Game length (turns): min %d, avg %.1f, max %d\n", low, avg, high)
	}
	testedWins := wins[tested-1]
	_, _ = fmt.Fprintf(w, "\nDeck under test win rate: %.1f%% (%d/%d)\n",
		percent(testedWins, result.GameCount), testedWins, result.GameCount)
	for _, failure := range result.Failures {
		_, _ = fmt.Fprintf(w, "  failed game %d (seed %d): %s\n", failure.Index, failure.Seed, failure.Reason)
	}
}

// simReport is the JSON report written for a simulation. It summarises the
// deck-under-test perspective rather than dumping every game log.
type simReport struct {
	Games       int           `json:"games"`
	MasterSeed  uint64        `json:"masterSeed"`
	Agent       string        `json:"agent"`
	TestedIndex int           `json:"testedIndex"`
	TestedDeck  string        `json:"testedDeck"`
	Decks       []string      `json:"decks"`
	WinsBySeat  []int         `json:"winsBySeat"`
	TestedWins  int           `json:"testedWins"`
	Draws       int           `json:"draws"`
	GameLength  lengthSummary `json:"gameLength"`
	Failures    []failureJSON `json:"failures,omitempty"`
}

type lengthSummary struct {
	Min     int     `json:"min"`
	Max     int     `json:"max"`
	Average float64 `json:"average"`
}

type failureJSON struct {
	Index  int    `json:"index"`
	Seed   uint64 `json:"seed"`
	Reason string `json:"reason"`
}

func writeSimReport(path string, result sim.SimulationResult, paths []string, tested int, profile string) error {
	wins := result.WinCounts()
	low, high, avg := gameLengthStats(result)
	report := simReport{
		Games:       result.GameCount,
		MasterSeed:  result.MasterSeed,
		Agent:       profileName(profile),
		TestedIndex: tested,
		TestedDeck:  deckName(paths[tested-1]),
		Decks:       deckNames(paths),
		WinsBySeat:  wins[:],
		TestedWins:  wins[tested-1],
		Draws:       result.DrawCount(),
		GameLength:  lengthSummary{Min: low, Max: high, Average: avg},
	}
	for _, failure := range result.Failures {
		report.Failures = append(report.Failures, failureJSON{Index: failure.Index, Seed: failure.Seed, Reason: failure.Reason})
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("encode report: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write report %q: %w", path, err)
	}
	return nil
}

// gameLengthStats returns the min, max, and average turn count over the games
// that completed (failed games are excluded). It returns zeros when none did.
func gameLengthStats(result sim.SimulationResult) (low, high int, avg float64) {
	failed := make(map[int]bool, len(result.Failures))
	for _, failure := range result.Failures {
		failed[failure.Index] = true
	}
	total, completed := 0, 0
	for i := range result.Games {
		if failed[i] {
			continue
		}
		turns := result.Games[i].TurnCount
		if completed == 0 || turns < low {
			low = turns
		}
		if turns > high {
			high = turns
		}
		total += turns
		completed++
	}
	if completed == 0 {
		return 0, 0, 0
	}
	return low, high, float64(total) / float64(completed)
}

func deckNames(paths []string) []string {
	names := make([]string, len(paths))
	for i, path := range paths {
		names[i] = deckName(path)
	}
	return names
}

func profileName(profile string) string {
	if profile == "" {
		return "firstlegal"
	}
	return profile
}

func percent(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return 100 * float64(part) / float64(total)
}
