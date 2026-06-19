package report

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/mtg/sim"
)

// CardMetrics is the per-card performance of one card name in the deck under
// test, aggregated across the completed games of a simulation. Counts cover only
// cards owned by the tested deck.
type CardMetrics struct {
	Name     string `json:"name"`
	Draws    int    `json:"draws"`
	Casts    int    `json:"casts"`
	Resolves int    `json:"resolves"`
	Discards int    `json:"discards"`
	Removed  int    `json:"removed"`
	// ZoneChanges is the total number of zone-change events for the card. The
	// engine emits a generic zone-change event for every move, so this is a
	// superset that already includes the card's draws, casts, discards, and
	// removals (plus other moves such as bounce, mill, exile, and tuck). It is a
	// coarse "how much this card moved around" figure and is not additive with
	// the columns above.
	ZoneChanges int `json:"zoneChanges"`
	// SeenInWins and SeenInLosses count the games (won or lost) in which the card
	// was drawn or cast at least once, so a card's record can be compared across
	// outcomes.
	SeenInWins   int `json:"seenInWins"`
	SeenInLosses int `json:"seenInLosses"`
	// Stranded counts how many times the card was left in the tested deck's hand
	// at game end, a sign it rotted rather than being played.
	Stranded int `json:"stranded"`
}

// computeCardMetrics aggregates per-card performance for the tested seat across
// the completed games of result, returning the cards sorted by casts (then draws,
// then name) so the order is stable and the most-played cards lead.
func computeCardMetrics(result sim.SimulationResult, seat game.PlayerID) []CardMetrics {
	failed := failedIndices(result)
	byName := make(map[string]*CardMetrics)
	metric := func(name string) *CardMetrics {
		existing, ok := byName[name]
		if !ok {
			existing = &CardMetrics{Name: name}
			byName[name] = existing
		}
		return existing
	}

	for i := range result.Games {
		if failed[i] {
			continue
		}
		gameResult := result.Games[i]
		won := gameResult.HasWinner && gameResult.Winner == seat
		seen := make(map[string]bool)

		for e := range gameResult.Events {
			event := gameResult.Events[e]
			name, ok := testedCardName(gameResult, event.CardID, seat)
			if !ok {
				continue
			}
			card := metric(name)
			switch event.Kind {
			case game.EventCardDrawn:
				card.Draws++
				seen[name] = true
			case game.EventSpellCast:
				card.Casts++
				seen[name] = true
			case game.EventSpellResolved:
				card.Resolves++
			case game.EventCardDiscarded:
				card.Discards++
			case game.EventPermanentDied:
				card.Removed++
			case game.EventZoneChanged:
				card.ZoneChanges++
			default:
			}
		}

		for _, cardID := range gameResult.EndState.Players[seat].Hand {
			if name, ok := testedCardName(gameResult, cardID, seat); ok {
				metric(name).Stranded++
			}
		}
		for name := range seen {
			if won {
				metric(name).SeenInWins++
			} else {
				metric(name).SeenInLosses++
			}
		}
	}

	return sortedCardMetrics(byName)
}

// testedCardName resolves a card instance to its name when it is owned by the
// tested seat, the filter that scopes every metric to the deck under test.
func testedCardName(result rules.GameResult, cardID id.ID, seat game.PlayerID) (string, bool) {
	info, ok := result.Cards[cardID]
	if !ok || info.Owner != seat {
		return "", false
	}
	return info.Name, true
}

func sortedCardMetrics(byName map[string]*CardMetrics) []CardMetrics {
	cards := make([]CardMetrics, 0, len(byName))
	for _, card := range byName {
		cards = append(cards, *card)
	}
	sort := func(a, b CardMetrics) int {
		if a.Casts != b.Casts {
			return cmp.Compare(b.Casts, a.Casts)
		}
		if a.Draws != b.Draws {
			return cmp.Compare(b.Draws, a.Draws)
		}
		return cmp.Compare(a.Name, b.Name)
	}
	slices.SortFunc(cards, sort)
	return cards
}

// writeCards renders the per-card table of the text summary, listing the most
// active cards first.
func writeCards(b *strings.Builder, cards []CardMetrics) {
	_, _ = fmt.Fprintf(b, "\nPer-card performance (%d cards):\n", len(cards))
	if len(cards) == 0 {
		return
	}
	_, _ = fmt.Fprintln(b, "  card                            draw cast resv disc rmvd strand  W/L")
	for i := range cards {
		card := cards[i]
		_, _ = fmt.Fprintf(b, "  %-30s %4d %4d %4d %4d %4d %6d  %d/%d\n",
			truncate(card.Name, 30), card.Draws, card.Casts, card.Resolves,
			card.Discards, card.Removed, card.Stranded, card.SeenInWins, card.SeenInLosses)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
