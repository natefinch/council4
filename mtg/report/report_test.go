package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/mtg/sim"
)

// syntheticResult builds a SimulationResult by hand — no engine, no live game —
// so the tests prove the report works purely from GameResult data.
func syntheticResult() sim.SimulationResult {
	return sim.SimulationResult{
		Games: []rules.GameResult{
			{HasWinner: true, Winner: game.Player1, TurnCount: 8},
			{HasWinner: true, Winner: game.Player2, TurnCount: 11},
			{}, // failed game: zero result
		},
		Seeds:      []uint64{10, 20, 30},
		GameCount:  3,
		MasterSeed: 99,
		Failures:   []sim.GameFailure{{Index: 2, Seed: 30, Reason: "unsupported card"}},
	}
}

func testOptions() Options {
	return Options{
		TestedSeat: game.Player1,
		DeckNames:  [game.NumPlayers]string{"Mine", "Opp A", "Opp B", "Opp C"},
	}
}

func TestGenerateSummarisesBatch(t *testing.T) {
	report := Generate(syntheticResult(), testOptions())

	if report.Games != 3 {
		t.Errorf("Games = %d, want 3", report.Games)
	}
	if report.Completed != 2 {
		t.Errorf("Completed = %d, want 2 (3 games - 1 failure)", report.Completed)
	}
	if report.TestedDeck != "Mine" {
		t.Errorf("TestedDeck = %q, want Mine", report.TestedDeck)
	}
	if report.TestedSeat != 0 {
		t.Errorf("TestedSeat = %d, want 0", report.TestedSeat)
	}
	if len(report.Failures) != 1 || report.Failures[0].Index != 2 {
		t.Errorf("Failures = %+v, want one failure at index 2", report.Failures)
	}
}

func TestWriteTextRendersSummary(t *testing.T) {
	report := Generate(syntheticResult(), testOptions())
	var out bytes.Buffer
	if err := report.WriteText(&out); err != nil {
		t.Fatalf("WriteText: %v", err)
	}
	text := out.String()
	for _, want := range []string{
		"Deck performance report",
		"Deck under test: Mine (seat 1)",
		"Games: 3 (2 completed, 1 failed)",
		"unsupported card",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("text summary missing %q\nfull text:\n%s", want, text)
		}
	}
}

func TestWriteJSONRoundTrips(t *testing.T) {
	report := Generate(syntheticResult(), testOptions())
	var out bytes.Buffer
	if err := report.WriteJSON(&out); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	var decoded Report
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if decoded.Games != report.Games || decoded.TestedDeck != report.TestedDeck {
		t.Errorf("round-tripped report = %+v, want %+v", decoded, report)
	}
	if len(decoded.DeckNames) != game.NumPlayers {
		t.Errorf("DeckNames = %d entries, want %d", len(decoded.DeckNames), game.NumPlayers)
	}
}
