package sim

import (
	"sync/atomic"
	"testing"

	"github.com/natefinch/council4/mtg/agent"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/rules"
)

// matchupSeatSpy is a SeatAgent-built agent that records which seat it was built
// for, so a matchup test can confirm the tested agent visited every seat. The
// counters are atomic because sim.Run calls the agent factory from parallel
// worker goroutines (a factory must be safe for concurrent use).
type matchupSeatSpy struct {
	agent.FirstLegal

	seat game.PlayerID
}

func recordingSeatAgent(seen *[game.NumPlayers]atomic.Int64) SeatAgent {
	return func(_ uint64, seat game.PlayerID) rules.PlayerAgent {
		seen[seat].Add(1)
		return matchupSeatSpy{seat: seat}
	}
}

func firstLegalSeatAgent() SeatAgent {
	return func(uint64, game.PlayerID) rules.PlayerAgent { return agent.FirstLegal{} }
}

func TestAccumulateRotationAttributesOutcomes(t *testing.T) {
	// Three games with the tested agent in seat 2: a tested win, a baseline win,
	// and a draw. accumulateRotation must tally each correctly.
	rotation := SimulationResult{
		Games: []rules.GameResult{
			{HasWinner: true, Winner: game.Player3}, // tested seat wins
			{HasWinner: true, Winner: game.Player1}, // a baseline seat wins
			{Termination: rules.TerminationDraw},
		},
	}
	var result MatchupResult
	accumulateRotation(&result, rotation, game.Player3)

	if result.TestedWins != 1 || result.BaselineWins != 1 || result.Draws != 1 {
		t.Fatalf("tallies = tested %d, baseline %d, draws %d; want 1/1/1",
			result.TestedWins, result.BaselineWins, result.Draws)
	}
	if result.TestedWinsBySeat[game.Player3] != 1 {
		t.Fatalf("TestedWinsBySeat[Player3] = %d, want 1", result.TestedWinsBySeat[game.Player3])
	}
}

func TestAccumulateRotationCountsFailuresNotDraws(t *testing.T) {
	rotation := SimulationResult{
		Games:    []rules.GameResult{{}, {HasWinner: true, Winner: game.Player1}},
		Failures: []GameFailure{{Index: 0}},
	}

	var result MatchupResult
	accumulateRotation(&result, rotation, game.Player1)

	if result.Failures != 1 {
		t.Fatalf("Failures = %d, want 1", result.Failures)
	}
	if result.Draws != 0 {
		t.Fatalf("Draws = %d, want 0 (the winner-less game was a failure)", result.Draws)
	}
	if result.TestedWins != 1 {
		t.Fatalf("TestedWins = %d, want 1", result.TestedWins)
	}
}

func TestAccumulateRotationSeparatesWorkBudgetFromExecutionFailures(t *testing.T) {
	rotation := SimulationResult{
		Games: []rules.GameResult{
			{Termination: rules.TerminationWorkBudget},
			{Termination: rules.TerminationUnsupported},
			{Termination: rules.TerminationTurnBudget, TurnLimitReached: true},
		},
		Failures: []GameFailure{{Index: 0}, {Index: 1, Unsupported: true}},
	}
	var result MatchupResult
	accumulateRotation(&result, rotation, game.Player1)
	if result.BudgetStops != 2 || result.Failures != 1 || result.Draws != 0 {
		t.Fatalf("budget/failure classifications: %+v", result)
	}
}

func TestRunMatchupRotatesTestedAgentThroughEverySeat(t *testing.T) {
	var seen [game.NumPlayers]atomic.Int64
	m := Matchup{
		Configs:      smokeConfigs(),
		GamesPerSeat: 2,
		Seed:         7,
		TurnLimit:    1,
		Tested:       recordingSeatAgent(&seen),
		Baseline:     firstLegalSeatAgent(),
	}
	RunMatchup(m)

	for seat := range seen {
		if seen[seat].Load() == 0 {
			t.Fatalf("tested agent never seated in seat %d; rotation missed a seat", seat)
		}
	}
}

func TestRunMatchupAccountingIsComplete(t *testing.T) {
	m := Matchup{
		Configs:      smokeConfigs(),
		GamesPerSeat: 3,
		Seed:         42,
		TurnLimit:    1,
		Tested:       firstLegalSeatAgent(),
		Baseline:     firstLegalSeatAgent(),
	}
	result := RunMatchup(m)

	if result.TotalGames != game.NumPlayers*3 {
		t.Fatalf("TotalGames = %d, want %d", result.TotalGames, game.NumPlayers*3)
	}
	sum := result.TestedWins + result.BaselineWins + result.Draws + result.Failures + result.BudgetStops
	if sum != result.TotalGames {
		t.Fatalf("outcomes sum to %d, want TotalGames %d", sum, result.TotalGames)
	}
	bySeat := 0
	for _, wins := range result.TestedWinsBySeat {
		bySeat += wins
	}
	if bySeat != result.TestedWins {
		t.Fatalf("TestedWinsBySeat sums to %d, want TestedWins %d", bySeat, result.TestedWins)
	}
}

func TestRunMatchupIsDeterministic(t *testing.T) {
	m := Matchup{
		Configs:      smokeConfig(1, 99).Configs,
		GamesPerSeat: 1,
		Seed:         99,
		Tested:       firstLegalSeatAgent(),
		Baseline:     firstLegalSeatAgent(),
	}
	first, second := RunMatchup(m), RunMatchup(m)
	if first != second {
		t.Fatal("RunMatchup is not deterministic for a fixed Matchup")
	}
	if first.TestedWins != 1 || first.BaselineWins != 3 || first.BudgetStops != 0 || first.Failures != 0 {
		t.Fatalf("each seat rotation should complete with a rules win: %+v", first)
	}
}

func TestTestedWinRate(t *testing.T) {
	r := MatchupResult{TotalGames: 8, TestedWins: 4, Failures: 0}
	if got := r.TestedWinRate(); got != 0.5 {
		t.Fatalf("win rate = %v, want 0.5", got)
	}
	// Failures are excluded from the denominator (completed games only).
	r = MatchupResult{TotalGames: 8, TestedWins: 3, Failures: 2}
	if got := r.TestedWinRate(); got != 0.5 {
		t.Fatalf("win rate with failures = %v, want 0.5 (3 of 6 completed)", got)
	}
	if got := (MatchupResult{}).TestedWinRate(); got != 0 {
		t.Fatalf("empty win rate = %v, want 0", got)
	}
}
