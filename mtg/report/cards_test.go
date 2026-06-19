package report

import (
	"bytes"
	"cmp"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/mtg/sim"
)

// cardFixture is a two-game simulation for the tested seat (Player1) with crafted
// events and end-states. Card IDs: 1 = "Bolt" (Player1), 2 = "Bear" (Player1),
// 3 = "Opp Card" (Player2, must be ignored).
func cardFixture() sim.SimulationResult {
	cards := map[id.ID]rules.CardInfo{
		1: {Name: "Bolt", Owner: game.Player1},
		2: {Name: "Bear", Owner: game.Player1},
		3: {Name: "Opp Card", Owner: game.Player2},
	}
	// Game 0: Player1 wins. Bolt is drawn, cast, resolves; Bear is drawn but left
	// stranded in hand. An opponent card is cast (must be ignored).
	game0 := rules.GameResult{
		HasWinner: true, Winner: game.Player1, TurnCount: 7,
		Events: []game.Event{
			{Kind: game.EventCardDrawn, CardID: 1},
			{Kind: game.EventSpellCast, CardID: 1},
			{Kind: game.EventSpellResolved, CardID: 1},
			{Kind: game.EventCardDrawn, CardID: 2},
			{Kind: game.EventSpellCast, CardID: 3}, // opponent's card, ignored
		},
		Cards: cards,
	}
	game0.EndState.Players[game.Player1].Hand = []id.ID{2} // Bear stranded

	// Game 1: Player1 loses. Bolt is drawn and discarded; Bear is drawn, cast,
	// and its permanent dies (removed).
	game1 := rules.GameResult{
		HasWinner: true, Winner: game.Player2, TurnCount: 9,
		Events: []game.Event{
			{Kind: game.EventCardDrawn, CardID: 1},
			{Kind: game.EventCardDiscarded, CardID: 1},
			{Kind: game.EventCardDrawn, CardID: 2},
			{Kind: game.EventSpellCast, CardID: 2},
			{Kind: game.EventPermanentDied, CardID: 2},
		},
		Cards: cards,
	}

	return sim.SimulationResult{
		Games:      []rules.GameResult{game0, game1},
		Seeds:      []uint64{1, 2},
		GameCount:  2,
		MasterSeed: 1,
	}
}

func cardByName(metrics []CardMetrics, name string) (CardMetrics, bool) {
	for _, m := range metrics {
		if m.Name == name {
			return m, true
		}
	}
	return CardMetrics{}, false
}

func TestComputeCardMetrics(t *testing.T) {
	metrics := computeCardMetrics(cardFixture(), game.Player1)

	if _, leaked := cardByName(metrics, "Opp Card"); leaked {
		t.Error("opponent's card leaked into the tested deck's per-card metrics")
	}

	bolt, ok := cardByName(metrics, "Bolt")
	if !ok {
		t.Fatal("Bolt missing from metrics")
	}
	if bolt.Draws != 2 || bolt.Casts != 1 || bolt.Resolves != 1 || bolt.Discards != 1 {
		t.Errorf("Bolt = %+v, want draws 2 casts 1 resolves 1 discards 1", bolt)
	}
	if bolt.SeenInWins != 1 || bolt.SeenInLosses != 1 {
		t.Errorf("Bolt seen W/L = %d/%d, want 1/1", bolt.SeenInWins, bolt.SeenInLosses)
	}

	bear, ok := cardByName(metrics, "Bear")
	if !ok {
		t.Fatal("Bear missing from metrics")
	}
	if bear.Draws != 2 || bear.Casts != 1 || bear.Removed != 1 {
		t.Errorf("Bear = %+v, want draws 2 casts 1 removed 1", bear)
	}
	if bear.Stranded != 1 {
		t.Errorf("Bear stranded = %d, want 1 (left in hand in game 0)", bear.Stranded)
	}
	// Bear was drawn in game 0 (a win) and drawn+cast in game 1 (a loss).
	if bear.SeenInWins != 1 || bear.SeenInLosses != 1 {
		t.Errorf("Bear seen W/L = %d/%d, want 1/1", bear.SeenInWins, bear.SeenInLosses)
	}
}

func TestCardMetricsSortedByActivityAndStable(t *testing.T) {
	metrics := computeCardMetrics(cardFixture(), game.Player1)
	if !slices.IsSortedFunc(metrics, func(a, b CardMetrics) int {
		if a.Casts != b.Casts {
			return cmp.Compare(b.Casts, a.Casts)
		}
		if a.Draws != b.Draws {
			return cmp.Compare(b.Draws, a.Draws)
		}
		return cmp.Compare(a.Name, b.Name)
	}) {
		t.Errorf("card metrics not sorted by casts/draws/name: %+v", metrics)
	}
}

func TestCardMetricsInReport(t *testing.T) {
	report := Generate(cardFixture(), Options{
		TestedSeat: game.Player1,
		DeckNames:  [game.NumPlayers]string{"Mine", "A", "B", "C"},
	})
	if len(report.Cards) != 2 {
		t.Fatalf("report.Cards = %d, want 2 (Bolt, Bear; opponent excluded)", len(report.Cards))
	}
	var out bytes.Buffer
	if err := report.WriteText(&out); err != nil {
		t.Fatalf("WriteText: %v", err)
	}
	text := out.String()
	if !strings.Contains(text, "Per-card performance") || !strings.Contains(text, "Bolt") || !strings.Contains(text, "Bear") {
		t.Errorf("text report missing per-card table or cards:\n%s", text)
	}
}
