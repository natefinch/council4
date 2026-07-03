package agent

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules"
)

// crackbackLethalPenalty is charged against an attack declaration that would
// leave the agent unable to survive the strongest opponent's next-turn swing —
// the "don't alpha-strike into a lethal crackback / hold creatures back as
// blockers" rule (playbook §8.3). It is large enough to outweigh the value of an
// over-committed attack, so the agent prefers a safer declaration (or holding
// back), but it is gated to genuine lethal exposure so it never makes the agent
// passive when attacking is safe.
const crackbackLethalPenalty = 80.0

// crackbackPenalty returns crackbackLethalPenalty when committing the declared
// attackers would leave the agent dead to the strongest single opponent's return
// attack, and zero otherwise. It models the worst realistic crackback: every one
// of an opponent's creatures untaps and swings, and the agent blocks with only
// the creatures still untapped after attacking (attackers without vigilance are
// tapped and cannot block).
func crackbackPenalty(obs rules.PlayerObservation, declare action.DeclareAttackersAction) float64 {
	me := obs.Player
	attacking := make(map[id.ID]bool, len(declare.Attackers))
	for i := range declare.Attackers {
		attacking[declare.Attackers[i].Attacker] = true
	}

	blockers := remainingBlockers(obs, me, attacking)
	worstCrackback := 0
	for _, creatures := range opponentCreatures(obs, me) {
		if unblocked := unblockedCrackback(creatures, blockers); unblocked > worstCrackback {
			worstCrackback = unblocked
		}
	}
	if obs.Life(me)-worstCrackback <= 0 {
		return crackbackLethalPenalty
	}
	return 0
}

// remainingBlockers lists the agent's creatures still untapped during opponents'
// turns after this attack: its untapped creatures, minus attackers that lack
// vigilance (which become tapped and cannot block the crackback).
func remainingBlockers(obs rules.PlayerObservation, me game.PlayerID, attacking map[id.ID]bool) []rules.PermanentView {
	var blockers []rules.PermanentView
	battlefield := obs.Battlefield()
	for i := range battlefield {
		creature := battlefield[i]
		if creature.Controller != me || creature.PhasedOut || creature.Tapped || !isCreaturePermanent(creature) {
			continue
		}
		if attacking[creature.ObjectID] && !creature.HasKeyword(game.Vigilance) {
			continue
		}
		blockers = append(blockers, creature)
	}
	return blockers
}

// opponentCreatures groups every opponent's creatures by controller. Every
// creature is counted because all of an opponent's creatures untap on their turn
// and could attack in the crackback.
func opponentCreatures(obs rules.PlayerObservation, me game.PlayerID) map[game.PlayerID][]rules.PermanentView {
	byOpponent := make(map[game.PlayerID][]rules.PermanentView)
	battlefield := obs.Battlefield()
	for i := range battlefield {
		creature := battlefield[i]
		if creature.Controller == me || creature.PhasedOut || !isCreaturePermanent(creature) {
			continue
		}
		byOpponent[creature.Controller] = append(byOpponent[creature.Controller], creature)
	}
	return byOpponent
}

// unblockedCrackback estimates the damage that gets through if the opponent
// attacks with all of creatures and the agent blocks optimally with blockers. It
// greedily assigns each blocker to the biggest attacker it can legally block —
// only a flyer or reacher can block a flyer — and sums the power of the attackers
// left unblocked. It is deliberately conservative: an attacker the agent cannot
// block always counts as full damage.
func unblockedCrackback(creatures, blockers []rules.PermanentView) int {
	attackers := append([]rules.PermanentView(nil), creatures...)
	slices.SortFunc(attackers, func(a, b rules.PermanentView) int { return max(0, b.Power) - max(0, a.Power) })

	available := append([]rules.PermanentView(nil), blockers...)
	total := 0
	for i := range attackers {
		if idx := pickBlocker(attackers[i], available); idx >= 0 {
			available = append(available[:idx], available[idx+1:]...)
			continue
		}
		total += max(0, attackers[i].Power)
	}
	return total
}

// pickBlocker returns the index of a blocker that can legally block attacker,
// preferring a non-flying blocker for a ground attacker so flyers stay free to
// block flyers. It returns -1 when no available blocker can block the attacker.
func pickBlocker(attacker rules.PermanentView, available []rules.PermanentView) int {
	needsEvasion := attacker.HasKeyword(game.Flying)
	fallback := -1
	for i := range available {
		blocker := available[i]
		canFly := blocker.HasKeyword(game.Flying) || blocker.HasKeyword(game.Reach)
		if needsEvasion && !canFly {
			continue
		}
		if !needsEvasion && canFly {
			if fallback < 0 {
				fallback = i
			}
			continue
		}
		return i
	}
	return fallback
}
