package sim

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/rules"
)

// SimulationResult is the structured outcome of a simulation run that the report
// layer consumes. It is the stable hand-off between the simulation harness and
// reporting: reporting depends on these fields, not on the runner internals.
//
// Retention: a SimulationResult keeps every game's full rules.GameResult (index
// i corresponds to game i and to Seeds[i]). Full results make reporting simple
// and self-contained at a memory cost of roughly N game logs. For very large N
// where that is too much, a caller can keep only the per-game seeds and
// reconstruct any game on demand with RunOne(cfg, i) / replay (see #588) instead
// of retaining every GameResult; the harness intentionally does the simple,
// fully-retained thing by default.
type SimulationResult struct {
	// Games holds every completed game's result in run order. Games[i] is the
	// game played with Seeds[i] and is reproducible via RunOne(cfg, i).
	Games []rules.GameResult
	// Seeds[i] is the per-game seed derived for game i (GameSeed(MasterSeed, i)).
	Seeds []uint64
	// GameCount is the number of games the run was configured to play.
	GameCount int
	// MasterSeed is the master seed every per-game seed derived from.
	MasterSeed uint64
	// Failures records games that could not complete normally. It is empty until
	// failure capture is wired in; a failed game still occupies its index in
	// Games (with a zero-value result) and is attributed here by index and seed.
	Failures []GameFailure
}

// GameFailure attributes a game that could not complete normally to its index
// and seed, with a human-readable reason. Failure capture populates it.
type GameFailure struct {
	Index  int
	Seed   uint64
	Reason string
}

// WinCounts tallies how many games each seat won. Games with no winner (a draw
// or a turn-cap stop) are not counted; see DrawCount.
func (r SimulationResult) WinCounts() [game.NumPlayers]int {
	var counts [game.NumPlayers]int
	for i := range r.Games {
		if r.Games[i].HasWinner {
			counts[r.Games[i].Winner]++
		}
	}
	return counts
}

// DrawCount is the number of completed games that ended with no winner.
func (r SimulationResult) DrawCount() int {
	draws := 0
	for i := range r.Games {
		if !r.Games[i].HasWinner {
			draws++
		}
	}
	return draws
}

// FailureCount is the number of games that could not complete normally.
func (r SimulationResult) FailureCount() int {
	return len(r.Failures)
}
