package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// buildTwoTurnEventLog returns a game whose event log holds previousTurn's
// events followed by currentTurn's events, partitioned by EventTurnStarts so the
// window accessors resolve the current and previous turns. The game's current
// turn number is 2.
func buildTwoTurnEventLog(previousTurn, currentTurn []game.Event) *game.Game {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.TurnNumber = 2
	g.EventTurnStarts = []int{0, len(previousTurn)}
	for _, event := range previousTurn {
		g.AppendEvent(event)
	}
	for _, event := range currentTurn {
		g.AppendEvent(event)
	}
	return g
}

// TestEventWindowCountSumAnyNextOrdinal exercises the shared aggregation
// operations and the player-vs-controller predicate distinction over one turn's
// window.
func TestEventWindowCountSumAnyNextOrdinal(t *testing.T) {
	g := buildTwoTurnEventLog(nil, []game.Event{
		{Kind: game.EventSpellCast, Controller: game.Player1},
		{Kind: game.EventSpellCast, Controller: game.Player1},
		{Kind: game.EventSpellCast, Controller: game.Player2},
		{Kind: game.EventLifeGained, Player: game.Player1, Amount: 3},
		{Kind: game.EventLifeGained, Player: game.Player1, Amount: 4},
		{Kind: game.EventLifeGained, Player: game.Player2, Amount: 9},
	})
	window := eventsThisTurnWindow(g)

	if got := window.count(eventKindController(game.EventSpellCast, game.Player1)); got != 2 {
		t.Fatalf("count(spell cast by P1) = %d, want 2", got)
	}
	if got := window.count(eventKindController(game.EventSpellCast, game.Player2)); got != 1 {
		t.Fatalf("count(spell cast by P2) = %d, want 1", got)
	}
	if got := window.nextOrdinal(eventKindController(game.EventSpellCast, game.Player1)); got != 3 {
		t.Fatalf("nextOrdinal(spell cast by P1) = %d, want 3", got)
	}
	if got := window.sumAmount(eventKindPlayer(game.EventLifeGained, game.Player1)); got != 7 {
		t.Fatalf("sumAmount(life gained by P1) = %d, want 7", got)
	}
	if !window.any(eventKindPlayer(game.EventLifeGained, game.Player2)) {
		t.Fatal("any(life gained by P2) = false, want true")
	}
	if window.any(eventKindPlayer(game.EventLifeLost, game.Player1)) {
		t.Fatal("any(life lost by P1) = true, want false")
	}
}

// TestEventWindowDistinctBatches confirms distinctBatches groups events that
// share a nonzero SimultaneousID into one occurrence while counting every
// zero-id event as its own occurrence (CR 701.8e batch semantics).
func TestEventWindowDistinctBatches(t *testing.T) {
	batch := g0IDGenBatch()
	g := buildTwoTurnEventLog(nil, []game.Event{
		{Kind: game.EventCardDiscarded, Player: game.Player1, SimultaneousID: batch},
		{Kind: game.EventCardDiscarded, Player: game.Player1, SimultaneousID: batch},
		{Kind: game.EventCardDiscarded, Player: game.Player1},
		{Kind: game.EventCardDiscarded, Player: game.Player2, SimultaneousID: batch},
	})
	window := eventsThisTurnWindow(g)

	if got := window.distinctBatches(eventKindPlayer(game.EventCardDiscarded, game.Player1)); got != 2 {
		t.Fatalf("distinctBatches(discard by P1) = %d, want 2 (one batch + one single)", got)
	}
	if got := window.distinctBatches(eventKindPlayer(game.EventCardDiscarded, game.Player2)); got != 1 {
		t.Fatalf("distinctBatches(discard by P2) = %d, want 1", got)
	}
}

func g0IDGenBatch() game.ObjectID {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	return g.IDGen.Next()
}

// TestEventWindowRespectsTurnBoundary confirms the current and previous turn
// windows partition the log at the recorded turn boundary, so per-turn
// aggregation resets each turn.
func TestEventWindowRespectsTurnBoundary(t *testing.T) {
	g := buildTwoTurnEventLog(
		[]game.Event{
			{Kind: game.EventSpellCast, Controller: game.Player1},
			{Kind: game.EventSpellCast, Controller: game.Player1},
		},
		[]game.Event{
			{Kind: game.EventSpellCast, Controller: game.Player1},
		},
	)

	if got := eventsThisTurnWindow(g).count(eventKindController(game.EventSpellCast, game.Player1)); got != 1 {
		t.Fatalf("current-turn spell casts = %d, want 1 (reset at turn boundary)", got)
	}
	if got := eventsPreviousTurnWindow(g).count(eventKindController(game.EventSpellCast, game.Player1)); got != 2 {
		t.Fatalf("previous-turn spell casts = %d, want 2", got)
	}
}

// TestEventsThisTurnWindowDoesNotClone proves the hot-path window accessor
// aliases the live event log rather than cloning it (as Game.EventsThisTurn
// does), and that aggregating through it allocates nothing.
func TestEventsThisTurnWindowDoesNotClone(t *testing.T) {
	g := buildTwoTurnEventLog(
		[]game.Event{{Kind: game.EventSpellCast, Controller: game.Player1}},
		[]game.Event{
			{Kind: game.EventSpellCast, Controller: game.Player1},
			{Kind: game.EventSpellCast, Controller: game.Player1},
		},
	)

	window := eventsThisTurnWindow(g)
	start := g.EventTurnStarts[g.Turn.TurnNumber-1]
	if len(window) == 0 {
		t.Fatal("current-turn window is empty")
	}
	if &window[0] != &g.Events[start] {
		t.Fatal("eventsThisTurnWindow returned a copy; it must alias the live event log")
	}

	pred := eventKindController(game.EventSpellCast, game.Player1)
	if allocs := testing.AllocsPerRun(100, func() {
		_ = eventsThisTurnWindow(g).count(pred)
	}); allocs != 0 {
		t.Fatalf("window count allocated %v times, want 0", allocs)
	}
}

func benchmarkEventLog() *game.Game {
	current := make([]game.Event, 0, 64)
	for range 64 {
		current = append(current, game.Event{Kind: game.EventSpellCast, Controller: game.Player1})
	}
	return buildTwoTurnEventLog(nil, current)
}

// BenchmarkEventWindowCount measures the non-cloning aggregation hot path; it
// should report zero allocations per op.
func BenchmarkEventWindowCount(b *testing.B) {
	g := benchmarkEventLog()
	pred := eventKindController(game.EventSpellCast, game.Player1)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = eventsThisTurnWindow(g).count(pred)
	}
}

// BenchmarkEventsThisTurnClone measures the cloning public accessor for
// comparison; it allocates a copy of the turn's events on every call.
func BenchmarkEventsThisTurnClone(b *testing.B) {
	g := benchmarkEventLog()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		count := 0
		for _, event := range g.EventsThisTurn() {
			if event.Kind == game.EventSpellCast && event.Controller == game.Player1 {
				count++
			}
		}
		_ = count
	}
}
