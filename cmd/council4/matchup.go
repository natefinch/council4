package main

import (
	"fmt"
	"io"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/sim"
)

// defaultMatchupTurnLimit caps each matchup game's length. A weak agent can drag
// a four-player game toward the engine's 1000-turn safety cap, which would make a
// measurement run take far too long, so games are stopped after this many turns
// and counted as draws. A real game with functioning win conditions ends well
// before this; the cap only bounds pathological durdling.
const defaultMatchupTurnLimit = 60

// runMatchup measures the -agent profile (tested) against the -baseline profile
// over the four decks, seating the tested agent in one rotating seat and the
// baseline in the others. It runs gamesPerSeat games in each of the four
// rotations so turn-order advantage cancels, then prints the tested agent's win
// rate against the 25% break-even a four-player free-for-all implies.
//
// For a pure agent-skill comparison, pass the same decklist four times so every
// seat plays a mirror; then the win-rate gap over 25% is purely agent strength.
// With four different decks the number also reflects the decks the tested agent
// happened to pilot.
func runMatchup(w io.Writer, paths []string, gamesPerSeat, workers int, seed uint64, testedProfile, baselineProfile string, registry *cards.Registry) error {
	if gamesPerSeat < 1 {
		return fmt.Errorf("-games must be at least 1 for a matchup, got %d", gamesPerSeat)
	}
	configs, err := loadConfigs(paths, 1, registry)
	if err != nil {
		return err
	}
	tested, err := seatAgentFor(testedProfile, configs)
	if err != nil {
		return fmt.Errorf("-agent: %w", err)
	}
	baseline, err := seatAgentFor(baselineProfile, configs)
	if err != nil {
		return fmt.Errorf("-baseline: %w", err)
	}

	result := sim.RunMatchup(sim.Matchup{
		Configs:      configs,
		GamesPerSeat: gamesPerSeat,
		Seed:         seed,
		Workers:      workers,
		TurnLimit:    defaultMatchupTurnLimit,
		Tested:       tested,
		Baseline:     baseline,
	})
	printMatchupSummary(w, result, paths, testedProfile, baselineProfile)
	return nil
}

// printMatchupSummary writes a human-readable matchup report: the run size, the
// tested agent's win rate against the break-even line, and the per-seat wins that
// expose turn-order effects.
func printMatchupSummary(w io.Writer, result sim.MatchupResult, paths []string, testedProfile, baselineProfile string) {
	breakEven := 1.0 / float64(game.NumPlayers)
	_, _ = fmt.Fprintf(w, "Matchup: %q (tested) vs %q (baseline)\n", profileName(testedProfile), profileName(baselineProfile))
	_, _ = fmt.Fprintf(w, "Decks: %v\n", deckNames(paths))
	_, _ = fmt.Fprintf(w, "Games: %d (%d per seat rotation x %d seats)\n",
		result.TotalGames, result.GamesPerSeat, game.NumPlayers)
	_, _ = fmt.Fprintf(w, "\nTested wins:   %d\n", result.TestedWins)
	_, _ = fmt.Fprintf(w, "Baseline wins: %d\n", result.BaselineWins)
	_, _ = fmt.Fprintf(w, "Draws:         %d\n", result.Draws)
	_, _ = fmt.Fprintf(w, "Failures:      %d\n", result.Failures)

	rate := result.TestedWinRate()
	_, _ = fmt.Fprintf(w, "\nTested win rate: %.1f%% (break-even is %.1f%%)\n", rate*100, breakEven*100)
	switch {
	case rate > breakEven:
		_, _ = fmt.Fprintf(w, "=> the tested agent is stronger than the baseline (+%.1f points)\n", (rate-breakEven)*100)
	case rate < breakEven:
		_, _ = fmt.Fprintf(w, "=> the tested agent is weaker than the baseline (%.1f points)\n", (rate-breakEven)*100)
	default:
		_, _ = fmt.Fprintln(w, "=> the tested agent matches the baseline")
	}

	_, _ = fmt.Fprintln(w, "\nTested wins by seat (turn order):")
	for seat, wins := range result.TestedWinsBySeat {
		_, _ = fmt.Fprintf(w, "  seat %d: %d\n", seat+1, wins)
	}
}
