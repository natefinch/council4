package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/mtg/sim"
)

// passSeat passes priority on every decision. It keeps the battlefield empty so
// smoke games end (by decking out) quickly, exercising the full simulation
// pipeline without the cost of long board-building games.
type passSeat struct{}

func (passSeat) ChooseAction(_ rules.PlayerObservation, _ []action.Action) action.Action {
	return action.Pass()
}

func passFactory(uint64) [game.NumPlayers]rules.PlayerAgent {
	var seated [game.NumPlayers]rules.PlayerAgent
	for i := range seated {
		seated[i] = passSeat{}
	}
	return seated
}

func smokePaths() []string {
	path := filepath.Join("testdata", "smoke.txt")
	return []string{path, path, path, path}
}

func TestSmokeFixtureLoadsIntoLegalConfigs(t *testing.T) {
	configs, err := loadConfigs(smokePaths(), 1, cards.NewDefaultRegistry())
	if err != nil {
		t.Fatalf("loadConfigs(smoke fixtures) = %v, want a clean load", err)
	}
	for seat := range configs {
		if configs[seat].Commander == nil {
			t.Errorf("seat %d has no commander", seat)
		}
		if len(configs[seat].Deck) != 99 {
			t.Errorf("seat %d deck = %d cards, want 99", seat, len(configs[seat].Deck))
		}
	}
}

func TestSimulationReportFromRealGames(t *testing.T) {
	paths := smokePaths()
	configs, err := loadConfigs(paths, 1, cards.NewDefaultRegistry())
	if err != nil {
		t.Fatalf("loadConfigs: %v", err)
	}
	result := sim.Run(sim.Config{Configs: configs, Games: 3, Seed: 7, NewAgents: passFactory})

	var out bytes.Buffer
	reportPath := filepath.Join(t.TempDir(), "report.json")
	if err := reportSimulation(&out, result, paths, 1, "firstlegal", reportPath); err != nil {
		t.Fatalf("reportSimulation: %v", err)
	}

	text := out.String()
	for _, want := range []string{"Simulation: 3 games", "Deck under test", "win rate", "Game length"} {
		if !strings.Contains(text, want) {
			t.Errorf("summary missing %q\nfull summary:\n%s", want, text)
		}
	}

	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	var report simReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	if report.Games != 3 {
		t.Errorf("report.Games = %d, want 3", report.Games)
	}
	if len(report.WinsBySeat) != game.NumPlayers {
		t.Fatalf("WinsBySeat = %d entries, want %d", len(report.WinsBySeat), game.NumPlayers)
	}
	total := report.Draws
	for _, w := range report.WinsBySeat {
		total += w
	}
	if total != report.Games {
		t.Errorf("wins (%v) + draws (%d) = %d, want %d games", report.WinsBySeat, report.Draws, total, report.Games)
	}
}

func TestReportExcludesFailuresFromDraws(t *testing.T) {
	// A batch with two wins, one genuine draw, and one failed game.
	result := sim.SimulationResult{
		Games: []rules.GameResult{
			{HasWinner: true, Winner: game.Player1, TurnCount: 10},
			{HasWinner: true, Winner: game.Player1, TurnCount: 12},
			{HasWinner: false, TurnCount: 1000}, // genuine draw / turn-cap stop
			{},                                  // failed game: zero, winner-less
		},
		Seeds:      []uint64{1, 2, 3, 4},
		GameCount:  4,
		MasterSeed: 1,
		Failures:   []sim.GameFailure{{Index: 3, Seed: 4, Reason: "boom"}},
	}

	var out bytes.Buffer
	reportPath := filepath.Join(t.TempDir(), "report.json")
	if err := reportSimulation(&out, result, smokePaths(), 1, "firstlegal", reportPath); err != nil {
		t.Fatalf("reportSimulation: %v", err)
	}

	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	var report simReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	if report.Draws != 1 {
		t.Errorf("report Draws = %d, want 1 (the failed game must not count as a draw)", report.Draws)
	}
	// wins + genuine draws + failures must account for every game exactly once.
	total := report.Draws + len(report.Failures)
	for _, w := range report.WinsBySeat {
		total += w
	}
	if total != report.Games {
		t.Errorf("wins+draws+failures = %d, want %d games", total, report.Games)
	}
	if !strings.Contains(out.String(), "Draws: 1") {
		t.Errorf("summary should report 1 draw, got:\n%s", out.String())
	}
}

// engineSeedGamma is sim's per-game stream constant (sim.seedGamma). Mirrored
// here because it is unexported, to guard against a random seat's RNG stream
// coinciding with the engine's.
const engineSeedGamma = 0x9e3779b97f4a7c15

func TestRandomSeatStreamsDoNotCollideWithEngine(t *testing.T) {
	for seat, stream := range seatStreams {
		if stream == engineSeedGamma {
			t.Errorf("seatStreams[%d] equals the engine stream constant; seat %d's RNG would mirror the engine", seat, seat)
		}
		for other := seat + 1; other < len(seatStreams); other++ {
			if seatStreams[other] == stream {
				t.Errorf("seatStreams[%d] and seatStreams[%d] are equal; seats would share an RNG stream", seat, other)
			}
		}
	}
}

func TestAgentFactoryProfiles(t *testing.T) {
	var configs [game.NumPlayers]game.PlayerConfig
	for _, profile := range []string{"", "firstlegal", "random", "generic", "search"} {
		factory, err := agentFactory(profile, configs)
		if err != nil {
			t.Fatalf("agentFactory(%q) = %v", profile, err)
		}
		seated := factory(123)
		for seat := range seated {
			if seated[seat] == nil {
				t.Errorf("profile %q seat %d agent is nil", profile, seat)
			}
		}
	}
	if _, err := agentFactory("bogus", configs); err == nil {
		t.Error("agentFactory(bogus) = nil error, want an error for an unknown profile")
	}
}

func TestRunDeckSimulationValidatesInput(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	if err := runDeckSimulation(io.Discard, smokePaths(), 1, 0, 0, 1, "firstlegal", "", registry); err == nil {
		t.Error("games=0 should be rejected")
	}
	if err := runDeckSimulation(io.Discard, smokePaths(), 1, 2, 0, 1, "bogus", "", registry); err == nil {
		t.Error("an unknown agent profile should be rejected")
	}
	if err := runDeckSimulation(io.Discard, []string{"only-one"}, 1, 2, 0, 1, "firstlegal", "", registry); err == nil {
		t.Error("the wrong number of deck paths should be rejected")
	}
}
