package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

// recordManaDevelopment fills the turn log's land-drop and available-mana
// figures for the active player. It is called once per turn, at the end of the
// active player's first precombat main phase, so the snapshot reflects the
// board after that turn's land drop.
func recordManaDevelopment(g *game.Game, log *TurnLog) {
	log.LandsPlayed = countLandsPlayed(log.Actions, log.ActivePlayer)
	log.ManaAvailable, log.ManaColors = availablePlayerMana(g, log.ActivePlayer)
}

// countLandsPlayed reports how many lands the given player played in the turn so
// far, derived from the recorded actions.
func countLandsPlayed(actions []ActionLog, playerID game.PlayerID) int {
	count := 0
	for i := range actions {
		entry := actions[i]
		if entry.Player == playerID && entry.Action.Kind == action.ActionPlayLand {
			count++
		}
	}
	return count
}

// availablePlayerMana counts the mana sources the player controls that can tap
// for mana this turn and the distinct colors they produce. Each source counts
// once, the same approximation the engine's own mana heuristic uses: the figure
// describes roughly how much mana is open, not the exact total. Phased-out
// permanents and summoning-sick mana dorks are excluded because they cannot tap;
// rituals are excluded because they are spells, not permanents.
func availablePlayerMana(g *game.Game, playerID game.PlayerID) (int, []string) {
	obs := NewObservation(g, playerID)
	battlefield := obs.Battlefield()
	total := 0
	var colors []string
	for i := range battlefield {
		permanent := battlefield[i]
		if permanent.Controller != playerID || !permanent.ProducesMana || permanent.PhasedOut {
			continue
		}
		if permanent.SummoningSick && slices.Contains(permanent.Types, types.Creature) {
			continue
		}
		total++
		for _, c := range permanent.ProducesColors {
			code := c.Abbreviation()
			if !slices.Contains(colors, code) {
				colors = append(colors, code)
			}
		}
	}
	return total, colors
}
