package sim

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/rules"
)

func TestSimulationResultCarriesSeedsAndCounts(t *testing.T) {
	cfg := smokeConfig(5, 808)
	result := Run(cfg)

	if result.GameCount != cfg.Games {
		t.Errorf("GameCount = %d, want %d", result.GameCount, cfg.Games)
	}
	if result.MasterSeed != cfg.Seed {
		t.Errorf("MasterSeed = %d, want %d", result.MasterSeed, cfg.Seed)
	}
	if len(result.Seeds) != cfg.Games || len(result.Games) != cfg.Games {
		t.Fatalf("Seeds/Games lengths = %d/%d, want %d", len(result.Seeds), len(result.Games), cfg.Games)
	}
	for i := range cfg.Games {
		if result.Seeds[i] != GameSeed(cfg.Seed, i) {
			t.Errorf("Seeds[%d] = %d, want %d", i, result.Seeds[i], GameSeed(cfg.Seed, i))
		}
	}
}

func TestSimulationResultAggregatesOutcomes(t *testing.T) {
	// Build a synthetic result so the aggregation is checked against a known
	// tally independent of how the smoke games happen to play out.
	result := SimulationResult{
		Games: []rules.GameResult{
			{HasWinner: true, Winner: game.Player1},
			{HasWinner: true, Winner: game.Player3},
			{HasWinner: true, Winner: game.Player1},
			{HasWinner: false}, // draw / turn-cap stop
		},
		Failures: []GameFailure{{Index: 9, Seed: 123, Reason: "boom"}},
	}

	wins := result.WinCounts()
	if wins[game.Player1] != 2 {
		t.Errorf("Player1 wins = %d, want 2", wins[game.Player1])
	}
	if wins[game.Player3] != 1 {
		t.Errorf("Player3 wins = %d, want 1", wins[game.Player3])
	}
	if wins[game.Player2] != 0 || wins[game.Player4] != 0 {
		t.Errorf("unexpected wins for Player2/Player4: %v", wins)
	}
	if result.DrawCount() != 1 {
		t.Errorf("DrawCount = %d, want 1", result.DrawCount())
	}
	if result.FailureCount() != 1 {
		t.Errorf("FailureCount = %d, want 1", result.FailureCount())
	}
}

func TestSimulationResultWinsPlusDrawsCoverEveryGame(t *testing.T) {
	result := Run(smokeConfig(12, 31415))
	wins := result.WinCounts()
	total := result.DrawCount()
	for _, n := range wins {
		total += n
	}
	if total != result.GameCount {
		t.Errorf("wins (%v) + draws (%d) = %d, want %d games", wins, result.DrawCount(), total, result.GameCount)
	}
}
