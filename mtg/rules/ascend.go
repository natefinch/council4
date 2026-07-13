package rules

import (
	"github.com/natefinch/council4/mtg/game"
)

// grantCityBlessing gives playerID the city's blessing (CR 702.131) if that
// player is in the game and does not already have it, emitting
// EventGotCityBlessing on the false->true transition so "whenever you get the
// city's blessing" triggers fire. The city's blessing is player-level persistent
// state that is never removed, so the event is emitted at most once per player
// per game. It reports whether the blessing was newly granted.
func grantCityBlessing(g *game.Game, playerID game.PlayerID) bool {
	player, ok := playerByID(g, playerID)
	if !ok || player.Eliminated || player.HasCityBlessing {
		return false
	}
	player.HasCityBlessing = true
	emitEvent(g, game.Event{
		Kind:       game.EventGotCityBlessing,
		Controller: playerID,
		Player:     playerID,
	})
	return true
}

// checkAscendCityBlessing applies the permanent form of ascend (CR 702.131b):
// "Any time you control ten or more permanents and you don't have the city's
// blessing, you get the city's blessing for the rest of the game." It scans the
// active rule effects for each RuleEffectAscend, and for every controller of one
// who controls ten or more permanents and lacks the blessing, grants it. This is
// a continuous check (not a triggered ability and it does not use the stack), so
// it runs as part of the state-based-action loop. It reports whether any player
// newly gained the city's blessing.
func checkAscendCityBlessing(g *game.Game) bool {
	var granted [game.NumPlayers]bool
	changed := false
	for _, effect := range activeRuleEffects(g) {
		if effect.Kind != game.RuleEffectAscend {
			continue
		}
		controller := effect.Controller
		if controller < 0 || int(controller) >= game.NumPlayers || granted[controller] {
			continue
		}
		player, ok := playerByID(g, controller)
		if !ok || player.HasCityBlessing {
			granted[controller] = true
			continue
		}
		if countControlledPermanents(g, controller) < 10 {
			continue
		}
		if grantCityBlessing(g, controller) {
			changed = true
		}
		granted[controller] = true
	}
	return changed
}

// countControlledPermanents counts the permanents the given player controls
// (CR 702.131). Only permanents on the battlefield count, phased-out permanents
// are ignored (CR 702.26e), and control is read through the continuous-effect
// layers so a permanent stolen by a control-changing effect counts for its
// current controller.
func countControlledPermanents(g *game.Game, controller game.PlayerID) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if activeBattlefieldPermanent(permanent) && effectiveController(g, permanent) == controller {
			count++
		}
	}
	return count
}

// handleGainCityBlessing resolves the GainCityBlessing primitive, the spell form
// of ascend (CR 702.131a): as the spell resolves, before its other instructions,
// its controller gets the city's blessing if they control ten or more permanents
// and don't already have it.
func handleGainCityBlessing(r *effectResolver, _ game.GainCityBlessing) effectResolved {
	res := effectResolved{accepted: true}
	controller := r.obj.Controller
	if countControlledPermanents(r.game, controller) >= 10 {
		res.succeeded = grantCityBlessing(r.game, controller)
	}
	return res
}
