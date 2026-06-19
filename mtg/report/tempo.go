package report

import (
	"fmt"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/mtg/sim"
)

// TempoMetrics describes how quickly the deck under test develops and how much
// damage it applies, across the completed games of a simulation.
type TempoMetrics struct {
	// ComeOnlineTurn is the average turn the deck cast its first spell, over the
	// games in which it cast at least one; OnlineGames is how many those were.
	ComeOnlineTurn float64 `json:"comeOnlineTurn"`
	OnlineGames    int     `json:"onlineGames"`
	// DamageDealtPerGame is the average combat damage the deck dealt to opponents
	// per game; DamagePerActiveTurn divides that by the deck's own turns.
	DamageDealtPerGame  float64 `json:"damageDealtPerGame"`
	DamagePerActiveTurn float64 `json:"damagePerActiveTurn"`
}

// CommanderMetrics describes the deck's reliance on its commander.
type CommanderMetrics struct {
	AverageCasts     float64     `json:"averageCasts"`
	CastDistribution map[int]int `json:"castDistribution"`
	// WinsWithCommander and WinsWithoutCommander split the deck's wins by whether
	// it had cast its commander; DependencyRate is the fraction of wins in which
	// the commander was cast — a high value means the deck leans on its commander.
	WinsWithCommander    int     `json:"winsWithCommander"`
	WinsWithoutCommander int     `json:"winsWithoutCommander"`
	DependencyRate       float64 `json:"dependencyRate"`
}

// InteractionMetrics counts opposing interaction aimed at the deck under test.
type InteractionMetrics struct {
	// TargetedByOpponents is how often an opponent's spell or ability targeted the
	// tested player or one of its permanents (a proxy for removal and disruption
	// aimed at the deck).
	TargetedByOpponents        int     `json:"targetedByOpponents"`
	TargetedByOpponentsPerGame float64 `json:"targetedByOpponentsPerGame"`
}

func computeTempo(result sim.SimulationResult, seat game.PlayerID) TempoMetrics {
	failed := failedIndices(result)
	onlineGames, onlineTurnSum := 0, 0
	totalDamage, totalTurns, completed := 0, 0, 0

	for i := range result.Games {
		if failed[i] {
			continue
		}
		completed++
		gameResult := result.Games[i]
		if turn, ok := comeOnlineTurn(gameResult, seat); ok {
			onlineGames++
			onlineTurnSum += turn
		}
		totalDamage += combatDamageDealt(gameResult, seat)
		totalTurns += testedTurnCount(gameResult, seat)
	}

	metrics := TempoMetrics{OnlineGames: onlineGames}
	if onlineGames > 0 {
		metrics.ComeOnlineTurn = float64(onlineTurnSum) / float64(onlineGames)
	}
	if completed > 0 {
		metrics.DamageDealtPerGame = float64(totalDamage) / float64(completed)
	}
	if totalTurns > 0 {
		metrics.DamagePerActiveTurn = float64(totalDamage) / float64(totalTurns)
	}
	return metrics
}

// comeOnlineTurn is the turn number on which the tested seat cast its first spell.
func comeOnlineTurn(result rules.GameResult, seat game.PlayerID) (int, bool) {
	for t := range result.Turns {
		turn := result.Turns[t]
		for a := range turn.Actions {
			entry := turn.Actions[a]
			if entry.Player == seat && entry.Action.Kind == action.ActionCastSpell {
				return turn.TurnNumber, true
			}
		}
	}
	return 0, false
}

func combatDamageDealt(result rules.GameResult, seat game.PlayerID) int {
	total := 0
	for t := range result.Turns {
		turn := result.Turns[t]
		for d := range turn.CombatDamage {
			damage := turn.CombatDamage[d]
			if damage.Controller == seat && damage.DefendingPlayer != seat {
				total += damage.Damage
			}
		}
	}
	return total
}

func testedTurnCount(result rules.GameResult, seat game.PlayerID) int {
	count := 0
	for t := range result.Turns {
		if result.Turns[t].ActivePlayer == seat {
			count++
		}
	}
	return count
}

func computeCommander(result sim.SimulationResult, seat game.PlayerID) CommanderMetrics {
	failed := failedIndices(result)
	distribution := make(map[int]int)
	totalCasts, completed := 0, 0
	winsWith, winsWithout := 0, 0

	for i := range result.Games {
		if failed[i] {
			continue
		}
		completed++
		gameResult := result.Games[i]
		casts := gameResult.EndState.Players[seat].CommanderCasts
		distribution[casts]++
		totalCasts += casts
		if gameResult.HasWinner && gameResult.Winner == seat {
			if casts > 0 {
				winsWith++
			} else {
				winsWithout++
			}
		}
	}

	metrics := CommanderMetrics{
		CastDistribution:     distribution,
		WinsWithCommander:    winsWith,
		WinsWithoutCommander: winsWithout,
	}
	if completed > 0 {
		metrics.AverageCasts = float64(totalCasts) / float64(completed)
	}
	if wins := winsWith + winsWithout; wins > 0 {
		metrics.DependencyRate = float64(winsWith) / float64(wins)
	}
	return metrics
}

func computeInteraction(result sim.SimulationResult, seat game.PlayerID) InteractionMetrics {
	failed := failedIndices(result)
	total, completed := 0, 0
	for i := range result.Games {
		if failed[i] {
			continue
		}
		completed++
		total += opponentTargets(result.Games[i], seat)
	}
	metrics := InteractionMetrics{TargetedByOpponents: total}
	if completed > 0 {
		metrics.TargetedByOpponentsPerGame = float64(total) / float64(completed)
	}
	return metrics
}

// opponentTargets counts the spells and abilities an opponent aimed at the tested
// player or its permanents. Permanents are attributed to an owner via the
// enter-the-battlefield events, which carry both the permanent and card IDs.
func opponentTargets(result rules.GameResult, seat game.PlayerID) int {
	permanentCard := permanentToCard(result)
	count := 0
	for e := range result.Events {
		event := result.Events[e]
		if event.Kind != game.EventObjectBecameTarget || event.Controller == seat {
			continue
		}
		switch event.Target.Kind {
		case game.TargetPlayer:
			if event.Target.PlayerID == seat {
				count++
			}
		case game.TargetPermanent:
			if cardID, ok := permanentCard[event.Target.PermanentID]; ok {
				if info, found := result.Cards[cardID]; found && info.Owner == seat {
					count++
				}
			}
		default:
		}
	}
	return count
}

func permanentToCard(result rules.GameResult) map[id.ID]id.ID {
	permanentCard := make(map[id.ID]id.ID)
	for e := range result.Events {
		event := result.Events[e]
		if event.Kind == game.EventPermanentEnteredBattlefield && event.PermanentID != 0 {
			permanentCard[event.PermanentID] = event.CardID
		}
	}
	return permanentCard
}

// writeTempo renders the tempo, commander, and interaction sections of the text
// summary.
func writeTempo(b *strings.Builder, tempo TempoMetrics, commander CommanderMetrics, interaction InteractionMetrics) {
	_, _ = fmt.Fprintln(b, "\nTempo:")
	if tempo.OnlineGames > 0 {
		_, _ = fmt.Fprintf(b, "  Comes online on turn %.1f (in %d games)\n", tempo.ComeOnlineTurn, tempo.OnlineGames)
	} else {
		_, _ = fmt.Fprintln(b, "  Comes online: n/a (no spells cast)")
	}
	_, _ = fmt.Fprintf(b, "  Combat damage dealt: %.1f per game (%.2f per active turn)\n",
		tempo.DamageDealtPerGame, tempo.DamagePerActiveTurn)

	_, _ = fmt.Fprintln(b, "\nCommander:")
	_, _ = fmt.Fprintf(b, "  Average casts per game: %.2f\n", commander.AverageCasts)
	_, _ = fmt.Fprintf(b, "  Wins with commander cast: %d; without: %d (dependency %.1f%%)\n",
		commander.WinsWithCommander, commander.WinsWithoutCommander, 100*commander.DependencyRate)

	_, _ = fmt.Fprintln(b, "\nOpponent interaction:")
	_, _ = fmt.Fprintf(b, "  Targeted by opponents: %d total (%.2f per game)\n",
		interaction.TargetedByOpponents, interaction.TargetedByOpponentsPerGame)
}
