package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// attackerLandwalkUnblockableBy reports whether the attacker's landwalk evasion
// (CR 702.14c) prevents blocker from blocking it: a creature with landwalk can't
// be blocked as long as the defending player controls a land matching the
// landwalk filter. The defending player is the controller of the prospective
// blocker.
func attackerLandwalkUnblockableBy(g *game.Game, attacker, blocker *game.Permanent) bool {
	values := effectivePermanentValues(g, attacker)
	if !values.keywords[game.Landwalk] {
		return false
	}
	defender := effectiveController(g, blocker)
	for i := range values.abilities {
		body, ok := values.abilities[i].(*game.StaticAbility)
		if !ok {
			continue
		}
		landwalk, ok := game.StaticBodyLandwalkKeyword(body)
		if !ok {
			continue
		}
		if defenderControlsLandwalkLand(g, defender, landwalk) {
			return true
		}
	}
	return false
}

// defenderControlsLandwalkLand reports whether defender controls a land matching
// the landwalk filter: any land for generic landwalk, or a land with the named
// subtype for a typed variant.
func defenderControlsLandwalkLand(g *game.Game, defender game.PlayerID, landwalk game.LandwalkKeyword) bool {
	for _, permanent := range g.Battlefield {
		if effectiveController(g, permanent) != defender {
			continue
		}
		if !permanentHasType(g, permanent, types.Land) {
			continue
		}
		if landwalk.AnyLand || permanentHasSubtype(g, permanent, landwalk.Subtype) {
			return true
		}
	}
	return false
}
