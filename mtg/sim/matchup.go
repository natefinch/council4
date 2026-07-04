package sim

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/rules"
)

// SeatAgent builds one seat's agent for one game. It receives the game's derived
// seed and the seat the agent will occupy, so agent randomness stays reproducible
// and an agent can depend on its seat. Give each seat its own RNG derived from
// gameSeed rather than sharing one.
type SeatAgent func(gameSeed uint64, seat game.PlayerID) rules.PlayerAgent

// Matchup measures one agent (Tested) against another (Baseline) over the same
// decks. It seats the tested agent in a single seat and the baseline in the other
// three, and rotates the tested seat through every position, running GamesPerSeat
// games in each rotation. Rotating the tested seat through all four positions
// cancels turn-order advantage, so if the two agents were equal strength the
// tested agent would win exactly one game in four (25%); a higher win rate means
// the tested agent is genuinely stronger.
//
// For a pure agent-skill measurement, seat the same deck in all four Configs (a
// mirror match): then the only differences between seats are turn order — which
// the rotation cancels — and agent skill. With four distinct decks the result
// mixes agent skill with deck strength across the seats the tested agent visits.
type Matchup struct {
	Configs      [game.NumPlayers]game.PlayerConfig
	GamesPerSeat int
	Seed         uint64
	Workers      int
	TurnLimit    int

	// Tested and Baseline build a seat's agent. Tested occupies one rotating
	// seat each game; Baseline fills the others.
	Tested   SeatAgent
	Baseline SeatAgent
}

// MatchupResult summarizes a Matchup. TestedWins plus BaselineWins plus Draws
// plus Failures equals TotalGames.
type MatchupResult struct {
	// GamesPerSeat and TotalGames record the run size; TotalGames is
	// GamesPerSeat times the number of seats.
	GamesPerSeat int
	TotalGames   int

	// TestedWins is the number of games the tested agent won; BaselineWins is the
	// number a baseline agent won. Draws are games that ended with no winner
	// (including turn-limit exhaustion); Failures are games that panicked.
	TestedWins   int
	BaselineWins int
	Draws        int
	Failures     int

	// TestedWinsBySeat[s] is how many games the tested agent won while seated in
	// seat s, so per-seat (turn-order) effects are visible.
	TestedWinsBySeat [game.NumPlayers]int
}

// TestedWinRate is the fraction of completed (non-failed) games the tested agent
// won. Compare it against 1/NumPlayers (0.25 in a four-player game): above that
// the tested agent is stronger than the baseline, below it weaker.
func (r MatchupResult) TestedWinRate() float64 {
	completed := r.TotalGames - r.Failures
	if completed <= 0 {
		return 0
	}
	return float64(r.TestedWins) / float64(completed)
}

// RunMatchup plays the matchup and returns its aggregated result. It runs
// GamesPerSeat games for each seat the tested agent can occupy, seating the
// tested agent there and the baseline elsewhere, and tallies wins by agent role.
// It is deterministic: the same Matchup yields the same result.
func RunMatchup(m Matchup) MatchupResult {
	result := MatchupResult{GamesPerSeat: m.GamesPerSeat}
	if m.GamesPerSeat <= 0 {
		return result
	}
	for testedSeat := range game.PlayerID(game.NumPlayers) {
		rotation := runMatchupRotation(m, testedSeat)
		result.TotalGames += rotation.GameCount
		accumulateRotation(&result, rotation, testedSeat)
	}
	return result
}

// runMatchupRotation runs the GamesPerSeat games with the tested agent in
// testedSeat. Each rotation gets its own seed stream (derived from the master
// seed and the seat) so the four rotations are distinct games rather than the
// same game replayed with the agent moved.
func runMatchupRotation(m Matchup, testedSeat game.PlayerID) SimulationResult {
	return Run(Config{
		Configs:   m.Configs,
		Games:     m.GamesPerSeat,
		Seed:      matchupRotationSeed(m.Seed, testedSeat),
		Workers:   m.Workers,
		TurnLimit: m.TurnLimit,
		NewAgents: func(gameSeed uint64) [game.NumPlayers]rules.PlayerAgent {
			var agents [game.NumPlayers]rules.PlayerAgent
			for seat := range agents {
				seatID := game.PlayerID(seat)
				build := m.Baseline
				if seatID == testedSeat {
					build = m.Tested
				}
				agents[seat] = build(gameSeed, seatID)
			}
			return agents
		},
	})
}

// accumulateRotation folds one rotation's games into the running result,
// attributing each game's outcome to the tested agent (when it won from
// testedSeat), a baseline agent, a draw, or a failure.
func accumulateRotation(result *MatchupResult, rotation SimulationResult, testedSeat game.PlayerID) {
	failed := failedGameIndices(rotation)
	for i := range rotation.Games {
		if failed[i] {
			result.Failures++
			continue
		}
		gameResult := rotation.Games[i]
		if !gameResult.HasWinner {
			result.Draws++
			continue
		}
		if gameResult.Winner == testedSeat {
			result.TestedWins++
			result.TestedWinsBySeat[testedSeat]++
			continue
		}
		result.BaselineWins++
	}
}

// failedGameIndices marks which games in a rotation panicked, so their winner-less
// results are counted as failures rather than draws.
func failedGameIndices(rotation SimulationResult) map[int]bool {
	failed := make(map[int]bool, len(rotation.Failures))
	for _, failure := range rotation.Failures {
		failed[failure.Index] = true
	}
	return failed
}

// matchupRotationSeed derives a distinct master seed for each rotation so no two
// rotations replay the same games. It mixes the seat into the seed with the same
// SplitMix64 construction used for per-game seeds. The seat guard keeps the
// conversion to uint64 provably non-negative.
func matchupRotationSeed(master uint64, testedSeat game.PlayerID) uint64 {
	seat := uint64(0)
	if testedSeat > 0 {
		seat = uint64(testedSeat)
	}
	return splitMix64(master + (seat+1)*seedGamma)
}
