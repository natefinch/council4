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
			// Actual rules draw; a turn cap is tested separately below.
			{Termination: rules.TerminationDraw, HasWinner: false, TurnCount: 1000,
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
	if outcome.Wins != 1 || outcome.Losses != 2 || outcome.Draws != 1 {
		t.Errorf("Wins/Losses/Draws = %d/%d/%d, want 1/2/1", outcome.Wins, outcome.Losses, outcome.Draws)
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
	if outcome.TurnsToLoss.Count != 2 || outcome.TurnsToLoss.Min != 10 || outcome.TurnsToLoss.Max != 12 {
		t.Errorf("TurnsToLoss = %+v, want min 10 max 12 over 2", outcome.TurnsToLoss)
	}
	if !approxEqual(outcome.TurnsToLoss.Average, 11) {
		t.Errorf("TurnsToLoss.Average = %v, want 11", outcome.TurnsToLoss.Average)
	}
}

func TestBudgetAndUnsupportedSlotsAreNotCompletedDraws(t *testing.T) {
	result := outcomeFixture()
	result.Games[3] = rules.GameResult{Termination: rules.TerminationTurnBudget, TurnLimitReached: true, TurnCount: 1000}
	result.Games[4].Termination = rules.TerminationUnsupported
	got := Generate(result, Options{})
	if got.Completed != 3 || got.Outcome.Completed != 3 || got.Outcome.Draws != 0 ||
		got.Terminations[rules.TerminationTurnBudget] != 1 || got.Terminations[rules.TerminationUnsupported] != 1 {
		t.Fatalf("budget/failure counted as a rules result: %+v", got)
	}

}

func TestPreviouslyEliminatedSeatLosesWhenSurvivorsDraw(t *testing.T) {
	result := outcomeFixture()
	got := computeOutcome(result, game.Player2)
	if got.Draws != 0 || got.Losses != 3 || got.Wins != 1 {
		t.Fatalf("previous loss was awarded a survivor's draw: %+v", got)
	}
}

func TestActualTerminalSimultaneousLossesReportDrawsAndTiedFinishes(t *testing.T) {
	engine := rules.NewEngine(nil)
	g := engine.NewGame([game.NumPlayers]game.PlayerConfig{})
	played := engine.RunGame(g, [game.NumPlayers]rules.PlayerAgent{})
	result := sim.SimulationResult{Games: []rules.GameResult{*played}, GameCount: 1}
	for player := range game.NumPlayers {
		outcome := computeOutcome(result, game.PlayerID(player))
		if outcome.Draws != 1 || outcome.Losses != 0 || outcome.TurnsToLoss.Count != 0 ||
			outcome.AverageFinish != 1 {
			t.Fatalf("terminal simultaneous loss not a draw/tie: %+v", outcome)
		}
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
