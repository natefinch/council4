package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
)

// playerUntargetableByRuleEffect reports whether a player-scoped hexproof or
// shroud rule effect blocks a source controlled by sourceController from
// targeting the given player. Shroud blocks every source (CR 702.18a); hexproof
// blocks only sources an opponent controls, so the player's own spells and
// abilities may still target them (CR 702.11b). A RuleEffectPlayerHexproof
// carrying Protection.FromColors is source-color-qualified "hexproof from those
// colors" (CR 702.11e): it blocks an opponent's source only when that source
// has one of the named colors. An empty FromColors is full hexproof.
//
// Unlike RuleEffectPlayerProtection, these player rule effects are purely a
// targeting restriction: they never prevent damage or other interactions, so
// they are consulted only on the targeting path (targetProtectedFromSource) and
// deliberately not by playerProtectedFromSource, which also drives damage
// prevention.
func playerUntargetableByRuleEffect(g *game.Game, sourceController, player game.PlayerID, sourceColors []color.Color) bool {
	for _, effect := range activeRuleEffects(g) {
		if !playerRelationMatches(effect.Controller, player, effect.AffectedPlayer) {
			continue
		}
		switch effect.Kind {
		case game.RuleEffectPlayerShroud:
			return true
		case game.RuleEffectPlayerHexproof:
			if sourceController == player {
				continue
			}
			from := effect.Protection.FromColors
			if len(from) == 0 || colorsIntersect(sourceColors, from) {
				return true
			}
		default:
			// Other rule effects impose no player-targeting restriction.
		}
	}
	return false
}

// colorsIntersect reports whether any color in a also appears in b.
func colorsIntersect(a, b []color.Color) bool {
	for _, c := range a {
		if slices.Contains(b, c) {
			return true
		}
	}
	return false
}
