// Package report turns simulation output into an actionable deck performance
// report: a human-readable text summary and a detailed JSON file about the deck
// under test.
//
// The report is produced purely from a sim.SimulationResult — the per-game
// rules.GameResult data, with the event stream and end-state folded in — so it
// never touches a live *game.Game. This keeps the report layer decoupled from
// the engine: the same input can be saved, replayed, and re-reported.
package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/sim"
)

// Options describes how to interpret a simulation for reporting: which seat held
// the deck under test and the display names of the four decks.
type Options struct {
	TestedSeat game.PlayerID
	DeckNames  [game.NumPlayers]string
}

// Report is the structured deck performance report. Later analysis layers extend
// it with outcome, per-card, mana, and interaction metrics; this is the stable
// envelope they populate, plus the text and JSON rendering.
type Report struct {
	Games      int       `json:"games"`
	MasterSeed uint64    `json:"masterSeed"`
	TestedSeat int       `json:"testedSeat"`
	TestedDeck string    `json:"testedDeck"`
	DeckNames  []string  `json:"deckNames"`
	Completed  int       `json:"completed"`
	Failures   []Failure `json:"failures,omitempty"`
}

// Failure attributes a game that could not complete to its index, seed, and
// reason.
type Failure struct {
	Index  int    `json:"index"`
	Seed   uint64 `json:"seed"`
	Reason string `json:"reason"`
}

// Generate builds a Report from a finished simulation and the reporting options.
// It reads only GameResult-derived data, never a live game.
func Generate(result sim.SimulationResult, opts Options) Report {
	report := Report{
		Games:      result.GameCount,
		MasterSeed: result.MasterSeed,
		TestedSeat: int(opts.TestedSeat),
		TestedDeck: opts.DeckNames[opts.TestedSeat],
		DeckNames:  opts.DeckNames[:],
		Completed:  result.GameCount - result.FailureCount(),
	}
	for _, failure := range result.Failures {
		report.Failures = append(report.Failures, Failure{
			Index:  failure.Index,
			Seed:   failure.Seed,
			Reason: failure.Reason,
		})
	}
	return report
}

// WriteText renders the human-readable summary to w.
func (r Report) WriteText(w io.Writer) error {
	var b strings.Builder
	_, _ = fmt.Fprintln(&b, "Deck performance report")
	_, _ = fmt.Fprintf(&b, "Deck under test: %s (seat %d)\n", r.TestedDeck, r.TestedSeat+1)
	_, _ = fmt.Fprintf(&b, "Games: %d (%d completed, %d failed)\n", r.Games, r.Completed, len(r.Failures))
	_, _ = fmt.Fprintf(&b, "Master seed: %d\n", r.MasterSeed)
	if len(r.Failures) > 0 {
		_, _ = fmt.Fprintln(&b, "\nFailed games:")
		for _, failure := range r.Failures {
			_, _ = fmt.Fprintf(&b, "  game %d (seed %d): %s\n", failure.Index, failure.Seed, failure.Reason)
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// WriteJSON renders the detailed report as indented JSON to w.
func (r Report) WriteJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(r)
}
