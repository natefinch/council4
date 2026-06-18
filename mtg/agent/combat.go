package agent

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/rules"
)

// Combat-scoring weights. Attacks are valued by the damage they apply
// (preferring pressure on the biggest threat) and by evasion and vigilance,
// and penalised when the defender can block profitably. Blocks are valued by
// the quality of the resulting trade.
const (
	attackDamageWeight   = 2.0
	attackThreatWeight   = 1.0
	attackEvasionBonus   = 3.0
	attackVigilanceBonus = 2.0
	attackLossBase       = 6.0

	blockKillAndSurvive = 12.0
	blockTradeBase      = 4.0
	blockTradePerPower  = 1.0
	blockChumpBase      = -3.0
	blockChumpPerPower  = 0.5
	blockPreventDamage  = 2.0
)

// scoreAttackDeclarations scores a declare-attackers action. Declaring no
// attackers is the score floor, so the agent attacks only when an attack has
// positive expected value.
func scoreAttackDeclarations(obs rules.PlayerObservation, act action.Action) float64 {
	declare, ok := act.DeclareAttackersPayload()
	if !ok || len(declare.Attackers) == 0 {
		return scorePass
	}
	model := NewThreatModel(obs)
	var total float64
	for i := range declare.Attackers {
		attacker, found := permanentByID(obs, declare.Attackers[i].Attacker)
		if !found {
			continue
		}
		total += attackValue(obs, model, attacker, declare.Attackers[i].Target)
	}
	return total
}

func attackValue(obs rules.PlayerObservation, model ThreatModel, attacker rules.PermanentView, target game.AttackTarget) float64 {
	power := max(0, attacker.Power)
	// An attacker the defender can profitably block likely dies for nothing, so
	// the attack is a loss scaled by the attacker we would forfeit.
	if defenderCanBlockProfitably(obs, attacker, target.Player) {
		return -(attackLossBase + float64(power))
	}
	value := attackDamageWeight * float64(power)
	value += attackThreatWeight * model.PlayerThreat(target.Player)
	if hasEvasion(attacker) {
		value += attackEvasionBonus
	}
	if attacker.HasKeyword(game.Vigilance) {
		value += attackVigilanceBonus
	}
	return value
}

// defenderCanBlockProfitably reports whether the defending player controls an
// untapped creature that could legally block attacker, kill it, and survive.
// Evasion is respected: a flyer can only be blocked by creatures with flying or
// reach, and a single blocker can never block a creature with menace.
func defenderCanBlockProfitably(obs rules.PlayerObservation, attacker rules.PermanentView, defender game.PlayerID) bool {
	if attacker.HasKeyword(game.Menace) {
		return false
	}
	battlefield := obs.Battlefield()
	for i := range battlefield {
		blocker := battlefield[i]
		if blocker.Controller != defender || blocker.Tapped || blocker.PhasedOut || !isCreaturePermanent(blocker) {
			continue
		}
		if !canBlock(attacker, blocker) {
			continue
		}
		if blockerKills(attacker, blocker) && !blockerDies(attacker, blocker) {
			return true
		}
	}
	return false
}

// canBlock reports whether blocker is able to legally block attacker given the
// attacker's evasion.
func canBlock(attacker, blocker rules.PermanentView) bool {
	if attacker.HasKeyword(game.Flying) {
		return blocker.HasKeyword(game.Flying) || blocker.HasKeyword(game.Reach)
	}
	return true
}

// scoreBlockDeclarations scores a declare-blockers action by the quality of each
// blocker-vs-attacker trade. Declaring no blockers is the score floor.
func scoreBlockDeclarations(obs rules.PlayerObservation, act action.Action) float64 {
	declare, ok := act.DeclareBlockersPayload()
	if !ok || len(declare.Blockers) == 0 {
		return scorePass
	}
	var total float64
	for i := range declare.Blockers {
		blocker, ok := permanentByID(obs, declare.Blockers[i].Blocker)
		if !ok {
			continue
		}
		attacker, ok := permanentByID(obs, declare.Blockers[i].Blocking)
		if !ok {
			continue
		}
		total += blockValue(attacker, blocker)
	}
	return total
}

func blockValue(attacker, blocker rules.PermanentView) float64 {
	power := float64(max(0, attacker.Power))
	kills := blockerKills(attacker, blocker)
	dies := blockerDies(attacker, blocker)
	switch {
	case kills && !dies:
		// Kill the attacker and keep the blocker: the best outcome. It scales
		// with the attacker's power so it always dominates merely absorbing the
		// same attacker.
		return blockKillAndSurvive + blockTradePerPower*power
	case kills && dies:
		// A trade: better when the attacker is bigger than the blocker.
		return blockTradeBase + blockTradePerPower*(power-float64(max(0, blocker.Power)))
	case !kills && dies:
		// A chump block: losing the blocker for nothing, worth it only to
		// prevent a large hit.
		return blockChumpBase + blockChumpPerPower*power
	default:
		// Absorb the attacker without losing the blocker.
		return blockPreventDamage + blockTradePerPower*power
	}
}

// blockerKills reports whether blocker deals lethal damage to attacker in
// combat, accounting for deathtouch and first strike.
func blockerKills(attacker, blocker rules.PermanentView) bool {
	if attacker.HasKeyword(game.FirstStrike) && !blocker.HasKeyword(game.FirstStrike) && attackerKills(attacker, blocker) {
		// The attacker strikes first and kills the blocker before it deals damage.
		return false
	}
	return dealsLethal(blocker, attacker)
}

// blockerDies reports whether attacker deals lethal damage to blocker in
// combat, accounting for deathtouch and first strike.
func blockerDies(attacker, blocker rules.PermanentView) bool {
	if blocker.HasKeyword(game.FirstStrike) && !attacker.HasKeyword(game.FirstStrike) && dealsLethal(blocker, attacker) {
		// The blocker strikes first and kills the attacker before it deals damage.
		return false
	}
	return attackerKills(attacker, blocker)
}

func attackerKills(attacker, blocker rules.PermanentView) bool {
	return dealsLethal(attacker, blocker)
}

// dealsLethal reports whether source deals lethal combat damage to target.
func dealsLethal(source, target rules.PermanentView) bool {
	power := max(0, source.Power)
	if power == 0 {
		return false
	}
	if source.HasKeyword(game.Deathtouch) {
		return true
	}
	return target.HasToughness && power >= target.Toughness
}

func hasEvasion(permanent rules.PermanentView) bool {
	return permanent.HasKeyword(game.Flying) || permanent.HasKeyword(game.Menace) || permanent.HasKeyword(game.Trample)
}
