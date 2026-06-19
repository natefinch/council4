package report

import (
	"math"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/mtg/sim"
)

// outcomeFixture is a fixed five-game simulation for the deck under test at seat
// Player1: a win, two losses at known positions, a draw the deck survived, and a
// failed game. It exercises every outcome path deterministically.
func outcomeFixture() sim.SimulationResult {
	return sim.SimulationResult{
		Games: []rules.GameResult{
			// Player1 wins.
			{HasWinner: true, Winner: game.Player1, TurnCount: 8,
				EliminationOrder: []game.PlayerID{game.Player2, game.Player3, game.Player4}},
			// Player1 eliminated first -> 4th place.
			{HasWinner: true, Winner: game.Player2, TurnCount: 12,
				EliminationOrder: []game.PlayerID{game.Player1, game.Player3, game.Player4}},
			// Player1 eliminated second -> 3rd place.
			{HasWinner: true, Winner: game.Player3, TurnCount: 10,
				EliminationOrder: []game.PlayerID{game.Player4, game.Player1, game.Player2}},
			// Draw at the turn cap; Player1 and Player4 survive -> tie for 1st.
			{HasWinner: false, TurnCount: 1000,
				EliminationOrder: []game.PlayerID{game.Player2, game.Player3}},
			// Failed game (zero result), excluded from all metrics.
			{},
		},
		Seeds:      []uint64{1, 2, 3, 4, 5},
		GameCount:  5,
		MasterSeed: 42,
		Failures:   []sim.GameFailure{{Index: 4, Seed: 5, Reason: "boom"}},
	}
}

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func TestComputeOutcomeMetrics(t *testing.T) {
	outcome := computeOutcome(outcomeFixture(), game.Player1)

	if outcome.Completed != 4 {
		t.Errorf("Completed = %d, want 4", outcome.Completed)
	}
	if outcome.Wins != 1 || outcome.Losses != 3 {
		t.Errorf("Wins/Losses = %d/%d, want 1/3", outcome.Wins, outcome.Losses)
	}
	if !approxEqual(outcome.WinRate, 0.25) {
		t.Errorf("WinRate = %v, want 0.25", outcome.WinRate)
	}
	// Finishes: game0 1st, game1 4th, game2 3rd, game3 1st (tie survivor).
	wantFinishes := []int{2, 0, 1, 1}
	for i, want := range wantFinishes {
		if outcome.FinishCounts[i] != want {
			t.Errorf("FinishCounts[%d] = %d, want %d (full: %v)", i, outcome.FinishCounts[i], want, outcome.FinishCounts)
		}
	}
	// Average finish: (1 + 4 + 3 + 1) / 4 = 2.25.
	if !approxEqual(outcome.AverageFinish, 2.25) {
		t.Errorf("AverageFinish = %v, want 2.25", outcome.AverageFinish)
	}
	// Game length over [8, 12, 10, 1000].
	if outcome.GameLength.Min != 8 || outcome.GameLength.Max != 1000 || !approxEqual(outcome.GameLength.Average, 257.5) {
		t.Errorf("GameLength = %+v, want min 8 max 1000 avg 257.5", outcome.GameLength)
	}
	// Turns to win over [8]; turns to loss over [12, 10, 1000].
	if outcome.TurnsToWin.Count != 1 || outcome.TurnsToWin.Min != 8 || outcome.TurnsToWin.Max != 8 {
		t.Errorf("TurnsToWin = %+v, want a single 8", outcome.TurnsToWin)
	}
	if outcome.TurnsToLoss.Count != 3 || outcome.TurnsToLoss.Min != 10 || outcome.TurnsToLoss.Max != 1000 {
		t.Errorf("TurnsToLoss = %+v, want min 10 max 1000 over 3", outcome.TurnsToLoss)
	}
	if !approxEqual(outcome.TurnsToLoss.Average, 1022.0/3.0) {
		t.Errorf("TurnsToLoss.Average = %v, want %v", outcome.TurnsToLoss.Average, 1022.0/3.0)
	}
}

func TestOutcomeAppearsInTextAndJSON(t *testing.T) {
	report := Generate(outcomeFixture(), Options{
		TestedSeat: game.Player1,
		DeckNames:  [game.NumPlayers]string{"Mine", "A", "B", "C"},
	})

	var out strings.Builder
	if err := report.WriteText(&out); err != nil {
		t.Fatalf("WriteText: %v", err)
	}
	text := out.String()
	for _, want := range []string{
		"Win rate: 25.0% (1/4)",
		"Average finishing position: 2.25",
		"Game length (turns): min 8, avg 257.5, max 1000",
		"Turns to win (turns): min 8",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("text missing %q\nfull:\n%s", want, text)
		}
	}

	if report.Outcome.Wins != 1 {
		t.Errorf("report.Outcome.Wins = %d, want 1", report.Outcome.Wins)
	}
}
