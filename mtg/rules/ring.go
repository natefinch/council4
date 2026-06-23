package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// ringMaxLevel is the highest level the Ring emblem reaches (CR 701.51). Once a
// player is at level 4, further temptings keep them there.
const ringMaxLevel = 4

// ringBearerLoseLifeOnDamage is the life each opponent loses when a level-4
// Ring-bearer deals combat damage to a player (the Ring's fourth ability).
const ringBearerLoseLifeOnDamage = 3

// handleRingTempts resolves the RingTempts primitive ("The Ring tempts you.",
// CR 701.51). The referenced player gets the Ring emblem if they don't have it
// and advances it to the next of its four levels, then chooses a creature they
// control to become (or remain) their Ring-bearer.
func handleRingTempts(r *effectResolver, prim game.RingTempts) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	player, ok := playerByID(r.game, playerID)
	if !ok || player.Eliminated {
		return res
	}
	advanceRingLevel(player)
	player.RingTemptedCount++
	r.designateRingBearer(player)
	res.succeeded = true
	return res
}

// advanceRingLevel raises the player's Ring emblem to its next level, capped at
// the fourth and final level.
func advanceRingLevel(player *game.Player) {
	if player.RingLevel < ringMaxLevel {
		player.RingLevel++
	}
}

// designateRingBearer asks the player to choose a creature they control to
// become (or remain) their Ring-bearer. If the player controls no creatures,
// any existing designation is left unchanged (CR 701.51c).
func (r *effectResolver) designateRingBearer(player *game.Player) {
	candidates := ringBearerCandidates(r.game, player.ID)
	if len(candidates) == 0 {
		return
	}
	options := make([]game.ChoiceOption, len(candidates))
	defaultIndex := 0
	for i, permanent := range candidates {
		options[i] = game.ChoiceOption{
			Index: i,
			Label: permanentEffectiveName(r.game, permanent),
			Card:  permanentChoiceInfo(r.game, permanent),
		}
		if permanent.ObjectID == player.RingBearerID {
			defaultIndex = i
		}
	}
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           player.ID,
		Prompt:           "Choose a creature to become your Ring-bearer",
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{defaultIndex},
	}, r.log)
	choice := defaultIndex
	if len(selected) == 1 && selected[0] >= 0 && selected[0] < len(candidates) {
		choice = selected[0]
	}
	player.RingBearerID = candidates[choice].ObjectID
}

// ringBearerCandidates returns the creatures the player controls, in battlefield
// order, that are eligible to become their Ring-bearer.
func ringBearerCandidates(g *game.Game, playerID game.PlayerID) []*game.Permanent {
	var candidates []*game.Permanent
	for _, permanent := range g.Battlefield {
		if effectiveController(g, permanent) == playerID && permanentHasType(g, permanent, types.Creature) {
			candidates = append(candidates, permanent)
		}
	}
	return candidates
}

// isRingBearer reports whether the permanent is the Ring-bearer of the player
// who controls it.
func isRingBearer(g *game.Game, permanent *game.Permanent) bool {
	controller, ok := playerByID(g, effectiveController(g, permanent))
	if !ok {
		return false
	}
	return controller.RingBearerID != id.ID(0) && controller.RingBearerID == permanent.ObjectID
}

// applyRingBearerCombatDamageToPlayer applies the Ring's fourth ability: when a
// level-4 player's Ring-bearer deals combat damage to a player, each of that
// player's opponents loses 3 life (CR 701.51, the Ring's fourth level).
func applyRingBearerCombatDamageToPlayer(g *game.Game, source *game.Permanent) {
	controllerID := effectiveController(g, source)
	controller, ok := playerByID(g, controllerID)
	if !ok || controller.RingLevel < ringMaxLevel || !isRingBearer(g, source) {
		return
	}
	for _, opponent := range g.Players {
		if opponent.ID == controllerID || opponent.Eliminated {
			continue
		}
		loseLife(g, opponent.ID, ringBearerLoseLifeOnDamage)
	}
}
