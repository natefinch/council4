package report

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/mtg/sim"
)

// Flood/screw classification thresholds. A flooded game played many lands but
// cast few spells; a screwed game had several turns yet played very few lands.
const (
	floodLandMin  = 6
	floodSpellMax = 2
	screwLandMax  = 2
	screwMinTurns = 4
)

// ManaCurveMetrics summarises how the deck under test developed and spent its
// mana across the completed games of a simulation.
type ManaCurveMetrics struct {
	// LandsPerTurn is the average lands the tested deck played per one of its
	// turns; ManaSpentPerGame and SpellsPerGame average mana value spent on
	// spells and the number of spells cast per game.
	LandsPerTurn     float64 `json:"landsPerTurn"`
	ManaSpentPerGame float64 `json:"manaSpentPerGame"`
	SpellsPerGame    float64 `json:"spellsPerGame"`
	// FloodRate and ScrewRate are the fraction of completed games that looked
	// flooded or screwed; NoLandDropRate is the fraction of the tested deck's
	// turns in which it played no land (a missed-land-drop proxy).
	FloodRate      float64 `json:"floodRate"`
	ScrewRate      float64 `json:"screwRate"`
	NoLandDropRate float64 `json:"noLandDropRate"`
	// RotMVPerGame is the average mana value of nonland cards drawn but never
	// cast, per game — expensive cards rotting in hand. RottedCards lists those
	// cards, most expensive first.
	RotMVPerGame float64      `json:"rotMVPerGame"`
	RottedCards  []RottedCard `json:"rottedCards,omitempty"`
}

// RottedCard is a nonland card the tested deck drew but never cast.
type RottedCard struct {
	Name      string `json:"name"`
	ManaValue int    `json:"manaValue"`
	Draws     int    `json:"draws"`
}

// computeManaCurve derives the mana and curve metrics for the tested seat across
// the completed games of result. cards supplies the per-card draw/cast tallies
// used to find rotted cards.
func computeManaCurve(result sim.SimulationResult, seat game.PlayerID, cards []CardMetrics) ManaCurveMetrics {
	failed := failedIndices(result)
	completed := 0
	totalLands, totalTurns, totalSpells, totalManaSpent, totalNoLandDrop := 0, 0, 0, 0, 0
	floodedGames, screwedGames := 0, 0

	for i := range result.Games {
		if failed[i] {
			continue
		}
		completed++
		stats := gameMana(result.Games[i], seat)
		totalLands += stats.lands
		totalTurns += stats.turns
		totalSpells += stats.spells
		totalManaSpent += stats.manaSpent
		totalNoLandDrop += stats.noLandDropTurns
		if stats.lands >= floodLandMin && stats.spells <= floodSpellMax {
			floodedGames++
		}
		if stats.turns >= screwMinTurns && stats.lands <= screwLandMax {
			screwedGames++
		}
	}

	metrics := ManaCurveMetrics{}
	if totalTurns > 0 {
		metrics.LandsPerTurn = float64(totalLands) / float64(totalTurns)
		metrics.NoLandDropRate = float64(totalNoLandDrop) / float64(totalTurns)
	}
	if completed > 0 {
		metrics.ManaSpentPerGame = float64(totalManaSpent) / float64(completed)
		metrics.SpellsPerGame = float64(totalSpells) / float64(completed)
		metrics.FloodRate = float64(floodedGames) / float64(completed)
		metrics.ScrewRate = float64(screwedGames) / float64(completed)
		metrics.RottedCards, metrics.RotMVPerGame = rottedCards(result, seat, cards, completed)
	}
	return metrics
}

type manaGameStats struct {
	lands, turns, spells, manaSpent, noLandDropTurns int
}

// gameMana tallies one game's land plays, tested-seat turns, spells cast, mana
// value spent, and turns the tested seat played no land.
func gameMana(result rules.GameResult, seat game.PlayerID) manaGameStats {
	var stats manaGameStats
	for t := range result.Turns {
		turn := result.Turns[t]
		if turn.ActivePlayer != seat {
			continue
		}
		stats.turns++
		landsThisTurn := 0
		for a := range turn.Actions {
			entry := turn.Actions[a]
			if entry.Player == seat && entry.Action.Kind == action.ActionPlayLand {
				landsThisTurn++
			}
		}
		stats.lands += landsThisTurn
		if landsThisTurn == 0 {
			stats.noLandDropTurns++
		}
	}
	for e := range result.Events {
		event := result.Events[e]
		if event.Kind != game.EventSpellCast {
			continue
		}
		if info, ok := result.Cards[event.CardID]; ok && info.Owner == seat {
			stats.spells++
			if event.ManaValue.Exists {
				stats.manaSpent += event.ManaValue.Val
			}
		}
	}
	return stats
}

// rottedCards returns the nonland cards the tested deck drew but never cast
// across the batch, most expensive first, plus the average rotted mana value per
// completed game.
func rottedCards(result sim.SimulationResult, seat game.PlayerID, cards []CardMetrics, completed int) (rotted []RottedCard, rotPerGame float64) {
	values, lands := cardValues(result, seat)
	totalRotMV := 0
	for i := range cards {
		card := cards[i]
		if card.Draws == 0 || card.Casts > 0 || lands[card.Name] {
			continue
		}
		manaValue := values[card.Name]
		rotted = append(rotted, RottedCard{Name: card.Name, ManaValue: manaValue, Draws: card.Draws})
		totalRotMV += manaValue * card.Draws
	}
	slices.SortFunc(rotted, func(a, b RottedCard) int {
		if a.ManaValue != b.ManaValue {
			return cmp.Compare(b.ManaValue, a.ManaValue)
		}
		return cmp.Compare(a.Name, b.Name)
	})
	if completed > 0 {
		rotPerGame = float64(totalRotMV) / float64(completed)
	}
	return rotted, rotPerGame
}

// cardValues maps each tested-owned card name to its mana value and whether it is
// a land, read from the folded card identities.
func cardValues(result sim.SimulationResult, seat game.PlayerID) (values map[string]int, lands map[string]bool) {
	values = make(map[string]int)
	lands = make(map[string]bool)
	for i := range result.Games {
		for _, info := range result.Games[i].Cards {
			if info.Owner != seat {
				continue
			}
			if _, seen := values[info.Name]; !seen {
				values[info.Name] = info.ManaValue
			}
			if slices.Contains(info.Types, types.Land) {
				lands[info.Name] = true
			}
		}
	}
	return values, lands
}

// writeManaCurve renders the mana and curve section of the text summary.
func writeManaCurve(b *strings.Builder, metrics ManaCurveMetrics) {
	_, _ = fmt.Fprintln(b, "\nMana & curve:")
	_, _ = fmt.Fprintf(b, "  Lands per turn: %.2f\n", metrics.LandsPerTurn)
	_, _ = fmt.Fprintf(b, "  Mana spent per game: %.1f over %.1f spells\n", metrics.ManaSpentPerGame, metrics.SpellsPerGame)
	_, _ = fmt.Fprintf(b, "  Flood rate: %.1f%%   Screw rate: %.1f%%   No-land-drop turns: %.1f%%\n",
		100*metrics.FloodRate, 100*metrics.ScrewRate, 100*metrics.NoLandDropRate)
	_, _ = fmt.Fprintf(b, "  Expensive rot: %.1f mana/game drawn but never cast\n", metrics.RotMVPerGame)
	for _, card := range metrics.RottedCards {
		_, _ = fmt.Fprintf(b, "    %s (mv %d): drawn %d, never cast\n", card.Name, card.ManaValue, card.Draws)
	}
}
