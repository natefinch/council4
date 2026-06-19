package report

import (
	"fmt"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/mtg/sim"
)

// OutcomeMetrics are the core outcome statistics for the deck under test across a
// simulation batch. They are computed over completed games only; failed games
// are excluded.
type OutcomeMetrics struct {
	Completed     int     `json:"completed"`
	Wins          int     `json:"wins"`
	Losses        int     `json:"losses"`
	WinRate       float64 `json:"winRate"`
	AverageFinish float64 `json:"averageFinish"`
	// FinishCounts[i] is how many games the deck finished in position i+1 (index
	// 0 is first place). Tied finishes share the better position (competition
	// ranking).
	FinishCounts []int `json:"finishCounts"`
	// GameLength is the turn-count distribution over all completed games;
	// TurnsToWin and TurnsToLoss are the same over games the deck won or lost.
	GameLength  LengthStats `json:"gameLength"`
	TurnsToWin  LengthStats `json:"turnsToWin"`
	TurnsToLoss LengthStats `json:"turnsToLoss"`
}

// LengthStats summarises a set of turn counts.
type LengthStats struct {
	Count     int         `json:"count"`
	Min       int         `json:"min"`
	Max       int         `json:"max"`
	Average   float64     `json:"average"`
	Histogram map[int]int `json:"histogram,omitempty"`
}

// computeOutcome derives the outcome metrics for seat across the completed games
// of result.
func computeOutcome(result sim.SimulationResult, seat game.PlayerID) OutcomeMetrics {
	failed := failedIndices(result)
	finishCounts := make([]int, game.NumPlayers)
	var lengths, winTurns, lossTurns []int
	wins, completed, finishSum := 0, 0, 0

	for i := range result.Games {
		if failed[i] {
			continue
		}
		gameResult := result.Games[i]
		completed++
		lengths = append(lengths, gameResult.TurnCount)
		position := finishingPosition(gameResult, seat)
		if position >= 1 && position <= len(finishCounts) {
			finishCounts[position-1]++
		}
		finishSum += position
		if gameResult.HasWinner && gameResult.Winner == seat {
			wins++
			winTurns = append(winTurns, gameResult.TurnCount)
		} else {
			lossTurns = append(lossTurns, gameResult.TurnCount)
		}
	}

	metrics := OutcomeMetrics{
		Completed:    completed,
		Wins:         wins,
		Losses:       completed - wins,
		FinishCounts: finishCounts,
		GameLength:   lengthStatsFrom(lengths),
		TurnsToWin:   lengthStatsFrom(winTurns),
		TurnsToLoss:  lengthStatsFrom(lossTurns),
	}
	if completed > 0 {
		metrics.WinRate = float64(wins) / float64(completed)
		metrics.AverageFinish = float64(finishSum) / float64(completed)
	}
	return metrics
}

// finishingPosition returns seat's 1-based finishing position in a game using
// competition ranking: position is one plus the number of seats that finished
// strictly ahead. The winner is ahead of everyone; a survivor is ahead of an
// eliminated seat; among eliminated seats the one eliminated later finished
// ahead; survivors of a draw tie.
func finishingPosition(result rules.GameResult, seat game.PlayerID) int {
	ahead := 0
	for other := range game.NumPlayers {
		otherSeat := game.PlayerID(other)
		if otherSeat == seat {
			continue
		}
		if placesAhead(result, otherSeat, seat) {
			ahead++
		}
	}
	return ahead + 1
}

func placesAhead(result rules.GameResult, a, b game.PlayerID) bool {
	if result.HasWinner {
		if a == result.Winner {
			return true
		}
		if b == result.Winner {
			return false
		}
	}
	aIndex, aEliminated := eliminationIndex(result, a)
	bIndex, bEliminated := eliminationIndex(result, b)
	switch {
	case aEliminated && bEliminated:
		return aIndex > bIndex // eliminated later places ahead
	case aEliminated && !bEliminated:
		return false // a died, b survived
	case !aEliminated && bEliminated:
		return true // a survived, b died
	default:
		return false // both survived: tie
	}
}

func eliminationIndex(result rules.GameResult, seat game.PlayerID) (int, bool) {
	for i, eliminated := range result.EliminationOrder {
		if eliminated == seat {
			return i, true
		}
	}
	return -1, false
}

func lengthStatsFrom(values []int) LengthStats {
	stats := LengthStats{Count: len(values)}
	if len(values) == 0 {
		return stats
	}
	histogram := make(map[int]int, len(values))
	total := 0
	for i, value := range values {
		if i == 0 || value < stats.Min {
			stats.Min = value
		}
		if value > stats.Max {
			stats.Max = value
		}
		total += value
		histogram[value]++
	}
	stats.Average = float64(total) / float64(len(values))
	stats.Histogram = histogram
	return stats
}

func failedIndices(result sim.SimulationResult) map[int]bool {
	failed := make(map[int]bool, len(result.Failures))
	for _, failure := range result.Failures {
		failed[failure.Index] = true
	}
	return failed
}

// writeOutcome renders the outcome section of the text summary.
func writeOutcome(b *strings.Builder, outcome OutcomeMetrics) {
	_, _ = fmt.Fprintf(b, "\nOutcome (over %d completed games):\n", outcome.Completed)
	_, _ = fmt.Fprintf(b, "  Win rate: %.1f%% (%d/%d)\n", 100*outcome.WinRate, outcome.Wins, outcome.Completed)
	_, _ = fmt.Fprintf(b, "  Average finishing position: %.2f\n", outcome.AverageFinish)
	_, _ = fmt.Fprint(b, "  Finishes:")
	for i, count := range outcome.FinishCounts {
		_, _ = fmt.Fprintf(b, " %s×%d", ordinal(i+1), count)
	}
	_, _ = fmt.Fprintln(b)
	writeLength(b, "Game length", outcome.GameLength)
	writeLength(b, "Turns to win", outcome.TurnsToWin)
	writeLength(b, "Turns to loss", outcome.TurnsToLoss)
}

func writeLength(b *strings.Builder, label string, stats LengthStats) {
	if stats.Count == 0 {
		_, _ = fmt.Fprintf(b, "  %s: n/a\n", label)
		return
	}
	_, _ = fmt.Fprintf(b, "  %s (turns): min %d, avg %.1f, max %d\n", label, stats.Min, stats.Average, stats.Max)
}

func ordinal(n int) string {
	switch n {
	case 1:
		return "1st"
	case 2:
		return "2nd"
	case 3:
		return "3rd"
	default:
		return fmt.Sprintf("%dth", n)
	}
}
