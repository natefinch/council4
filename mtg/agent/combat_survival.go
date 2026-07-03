package agent

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules"
)

// commanderDamageLethal is the commander damage total at which a player loses
// (CR 903.10a): 21 damage from a single commander.
const commanderDamageLethal = 21

// survivalValue rewards a block assignment that keeps the agent alive. If taking
// the current attack with the proposed blocks would leave the agent dead — by
// life total or by commander damage — the assignment earns nothing; if the
// unblocked attack is lethal but the proposed blocks avert it, the assignment
// earns blockSurvival, which dominates trade scoring so the agent chump-blocks to
// survive. When no attack threatens lethal, survival is neutral and ordinary
// trade quality decides.
func survivalValue(obs rules.PlayerObservation, blocks []game.BlockDeclaration) float64 {
	me := obs.Player
	attackers := obs.AttackersAgainst(me)
	if len(attackers) == 0 {
		return 0
	}
	lethalUnblocked := attackIsLethal(obs, attackers, nil)
	if !lethalUnblocked {
		return 0
	}
	if attackIsLethal(obs, attackers, blocks) {
		// These blocks do not save the agent, so they earn no survival credit —
		// the assignment is judged on trades alone.
		return 0
	}
	return blockSurvival
}

// attackIsLethal reports whether the attackers, with the given blocks applied,
// would kill the agent this combat — either by reducing its life to zero or less
// or by pushing commander damage from a single commander to 21 or more.
func attackIsLethal(obs rules.PlayerObservation, attackers []rules.AttackerView, blocks []game.BlockDeclaration) bool {
	me := obs.Player
	blockedToughness, blocked := blockAssignment(obs, blocks)

	faceDamage := 0
	commanderDamage := obs.PlayerState(me).CommanderDamage
	for i := range attackers {
		attacker := attackers[i].Attacker
		toFace := faceDamageFrom(attacker, blocked[attacker.ObjectID], blockedToughness[attacker.ObjectID])
		if toFace <= 0 {
			continue
		}
		faceDamage += toFace
		if commanderInstance(obs, attacker.CardInstanceID) {
			if commanderDamage[attacker.CardInstanceID]+toFace >= commanderDamageLethal {
				return true
			}
		}
	}
	return obs.Life(me)-faceDamage <= 0
}

// faceDamageFrom returns the damage an attacker deals to the defending player:
// its full power when unblocked, its trample overflow past its blockers' total
// toughness when blocked with trample, and nothing when blocked without trample.
func faceDamageFrom(attacker rules.PermanentView, isBlocked bool, blockersToughness int) int {
	power := max(0, attacker.Power)
	if !isBlocked {
		return power
	}
	if attacker.HasKeyword(game.Trample) {
		if overflow := power - blockersToughness; overflow > 0 {
			return overflow
		}
	}
	return 0
}

// blockAssignment summarizes a proposed block declaration: which attackers are
// blocked, and the total toughness of the blockers assigned to each, so trample
// overflow can be computed.
func blockAssignment(obs rules.PlayerObservation, blocks []game.BlockDeclaration) (toughnessByAttacker map[id.ID]int, blocked map[id.ID]bool) {
	toughnessByAttacker = make(map[id.ID]int, len(blocks))
	blocked = make(map[id.ID]bool, len(blocks))
	for i := range blocks {
		blocked[blocks[i].Blocking] = true
		if blocker, ok := permanentByID(obs, blocks[i].Blocker); ok && blocker.HasToughness {
			toughnessByAttacker[blocks[i].Blocking] += max(0, blocker.Toughness)
		}
	}
	return toughnessByAttacker, blocked
}

// commanderInstance reports whether the given card instance is a commander of any
// player, so an attacker that is a commander can be checked against the 21-damage
// commander-damage loss condition.
func commanderInstance(obs rules.PlayerObservation, cardInstanceID id.ID) bool {
	if cardInstanceID == 0 {
		return false
	}
	players := obs.Players()
	for i := range players {
		if players[i].CommanderInstanceID == cardInstanceID {
			return true
		}
	}
	return false
}
