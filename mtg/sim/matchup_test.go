package sim

import (
	"testing"

	"github.com/natefinch/council4/mtg/agent"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/rules"
)

// matchupSeatSpy is a SeatAgent-built agent that records which seat it was built
// for, so a matchup test can confirm the tested agent visited every seat.
type matchupSeatSpy struct {
	agent.FirstLegal

	seat game.PlayerID
	seen *[game.NumPlayers]int
}

func recordingSeatAgent(seen *[game.NumPlayers]int) SeatAgent {
	return func(_ uint64, seat game.PlayerID) rules.PlayerAgent {
		seen[seat]++
		return matchupSeatSpy{seat: seat, seen: seen}
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
			{HasWinner: false},                      // draw
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

func TestRunMatchupRotatesTestedAgentThroughEverySeat(t *testing.T) {
	var seen [game.NumPlayers]int
	m := Matchup{
		Configs:      smokeConfigs(),
		GamesPerSeat: 2,
		Seed:         7,
		TurnLimit:    12,
		Tested:       recordingSeatAgent(&seen),
		Baseline:     firstLegalSeatAgent(),
	}
	RunMatchup(m)

	for seat := range seen {
		if seen[seat] == 0 {
			t.Fatalf("tested agent never seated in seat %d; rotation missed a seat: %v", seat, seen)
		}
	}
}

func TestRunMatchupAccountingIsComplete(t *testing.T) {
	m := Matchup{
		Configs:      smokeConfigs(),
		GamesPerSeat: 3,
		Seed:         42,
		TurnLimit:    12,
		Tested:       firstLegalSeatAgent(),
		Baseline:     firstLegalSeatAgent(),
	}
	result := RunMatchup(m)

	if result.TotalGames != game.NumPlayers*3 {
		t.Fatalf("TotalGames = %d, want %d", result.TotalGames, game.NumPlayers*3)
	}
	sum := result.TestedWins + result.BaselineWins + result.Draws + result.Failures
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
		Configs:      smokeConfigs(),
		GamesPerSeat: 2,
		Seed:         99,
		TurnLimit:    12,
		Tested:       firstLegalSeatAgent(),
		Baseline:     firstLegalSeatAgent(),
	}
	if first, second := RunMatchup(m), RunMatchup(m); first != second {
		t.Fatal("RunMatchup is not deterministic for a fixed Matchup")
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
