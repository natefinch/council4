package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

// eventWindow is a read-only view over one turn's slice of the game event log.
// It is the single per-turn aggregation surface for conditions, dynamic
// amounts, limits, and trigger ordinals. Unlike Game.EventsThisTurn it does not
// clone the underlying events, so hot-path queries allocate nothing; callers
// must treat the events as read-only.
type eventWindow []game.Event

// eventsThisTurnWindow returns the current turn's events as a non-cloning
// window for read-only aggregation.
func eventsThisTurnWindow(g *game.Game) eventWindow {
	return eventsForTurnWindow(g, g.Turn.TurnNumber)
}

// eventsPreviousTurnWindow returns the previous turn's events as a non-cloning
// window for read-only aggregation.
func eventsPreviousTurnWindow(g *game.Game) eventWindow {
	return eventsForTurnWindow(g, g.Turn.TurnNumber-1)
}

// eventsForTurnWindow returns the events emitted during turnNumber as a
// non-cloning sub-slice of g.Events, using the same turn partition and bounds
// as Game.EventsForTurn but without copying. It returns nil for turns outside
// the recorded range.
func eventsForTurnWindow(g *game.Game, turnNumber int) eventWindow {
	if turnNumber <= 0 {
		return nil
	}
	index := turnNumber - 1
	if index < 0 || index >= len(g.EventTurnStarts) {
		return nil
	}
	start := g.EventTurnStarts[index]
	end := len(g.Events)
	if index+1 < len(g.EventTurnStarts) {
		end = g.EventTurnStarts[index+1]
	}
	if start < 0 || start > end || end > len(g.Events) {
		return nil
	}
	return eventWindow(g.Events[start:end])
}

// eventPredicate reports whether an event satisfies an aggregation filter.
type eventPredicate func(game.Event) bool

// eventKindPlayer matches events of kind whose Player field is player. It keeps
// the player-versus-controller distinction explicit: use it for player-relative
// queries (draws, discards, cycles, life changes).
func eventKindPlayer(kind game.EventKind, player game.PlayerID) eventPredicate {
	return func(event game.Event) bool {
		return event.Kind == kind && event.Player == player
	}
}

// eventKindController matches events of kind whose Controller field is
// controller. Use it for controller-relative queries (spells cast).
func eventKindController(kind game.EventKind, controller game.PlayerID) eventPredicate {
	return func(event game.Event) bool {
		return event.Kind == kind && event.Controller == controller
	}
}

// count returns how many events in the window satisfy match.
func (w eventWindow) count(match eventPredicate) int {
	count := 0
	for i := range w {
		if match(w[i]) {
			count++
		}
	}
	return count
}

// sumAmount returns the total Amount of the events in the window that satisfy
// match.
func (w eventWindow) sumAmount(match eventPredicate) int {
	total := 0
	for i := range w {
		if match(w[i]) {
			total += w[i].Amount
		}
	}
	return total
}

// any reports whether any event in the window satisfies match.
func (w eventWindow) any(match eventPredicate) bool {
	return slices.ContainsFunc(w, match)
}

// nextOrdinal returns the 1-based ordinal the next matching event would take:
// one more than the matches already in the window. It implements the "Nth ...
// each turn" ordinal (CR 700.6) for the next event of a kind.
func (w eventWindow) nextOrdinal(match eventPredicate) int {
	return w.count(match) + 1
}

// distinctBatches counts the distinct simultaneous occurrences among the events
// satisfying match. Events sharing one nonzero SimultaneousID form a single
// occurrence; an event with SimultaneousID 0 is its own occurrence (CR 701.8e,
// "the first time you discard one or more cards each turn").
func (w eventWindow) distinctBatches(match eventPredicate) int {
	batches := 0
	var seen map[id.ID]bool
	for i := range w {
		if !match(w[i]) {
			continue
		}
		simultaneousID := w[i].SimultaneousID
		if simultaneousID == 0 {
			batches++
			continue
		}
		if seen == nil {
			seen = make(map[id.ID]bool)
		}
		if !seen[simultaneousID] {
			seen[simultaneousID] = true
			batches++
		}
	}
	return batches
}
