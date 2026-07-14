package rules

import "github.com/natefinch/council4/mtg/game"

// playerUntargetableByRuleEffect reports whether a player-scoped hexproof or
// shroud rule effect blocks a source controlled by sourceController from
// targeting the given player. Shroud blocks every source (CR 702.18a); hexproof
// blocks only sources an opponent controls, so the player's own spells and
// abilities may still target them (CR 702.11b).
//
// Unlike RuleEffectPlayerProtection, these player rule effects are purely a
// targeting restriction: they never prevent damage or other interactions, so
// they are consulted only on the targeting path (targetProtectedFromSource) and
// deliberately not by playerProtectedFromSource, which also drives damage
// prevention.
func playerUntargetableByRuleEffect(g *game.Game, sourceController, player game.PlayerID) bool {
	for _, effect := range activeRuleEffects(g) {
		if !playerRelationMatches(effect.Controller, player, effect.AffectedPlayer) {
			continue
		}
		switch effect.Kind {
		case game.RuleEffectPlayerShroud:
			return true
		case game.RuleEffectPlayerHexproof:
			if sourceController != player {
				return true
			}
		default:
			// Other rule effects impose no player-targeting restriction.
		}
	}
	return false
}
