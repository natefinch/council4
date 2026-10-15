package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

type blockDeclarationProblem struct {
	game       *game.Game
	player     game.PlayerID
	attackers  []*game.Permanent
	blockers   []*game.Permanent
	allowed    [][]bool
	lures      [][]int
	minimum    []int
	oneBlocker []bool
	weights    []int
	capacities []int
	unit       bool
	states     int
}

func newBlockDeclarationProblem(g *game.Game, player game.PlayerID) *blockDeclarationProblem {
	g.BeginStaticSourceFrame()
	defer g.EndStaticSourceFrame()
	p := &blockDeclarationProblem{game: g, player: player, blockers: eligibleBlockers(g, player), unit: true}
	effects := activeRuleEffects(g)
	for _, declaration := range attacksAgainstPlayer(g, player) {
		attacker, exists := permanentByObjectID(g, declaration.Attacker)
		if !exists {
			continue
		}
		p.attackers = append(p.attackers, attacker)
		minimum, weight := 1, 0
		if attackerRequiresMultipleBlockers(g, attacker) {
			minimum, p.unit = 2, false
		}
		for index := range effects {
			if effects[index].Kind == game.RuleEffectMustBeBlocked && ruleEffectMatchesPermanent(g, &effects[index], attacker) {
				weight++
			}
		}
		if weight > 1 {
			p.unit = false
		}
		p.minimum, p.weights = append(p.minimum, minimum), append(p.weights, weight)
		p.oneBlocker = append(p.oneBlocker, ruleEffectLimitsBlockersToOne(g, attacker))
	}
	for _, blocker := range p.blockers {
		p.capacities = append(p.capacities, blockerBlockLimit(g, blocker))
		allowed, lures := make([]bool, len(p.attackers)), make([]int, len(p.attackers))
		if ruleEffectProhibitsBlockingAlone(g, blocker) {
			p.unit = false
		}
		for index, attacker := range p.attackers {
			allowed[index] = canBlockAttacker(g, blocker, attacker)
			if !allowed[index] {
				continue
			}
			for effect := range effects {
				if effects[effect].Kind == game.RuleEffectMustBeBlockedByAllAble && ruleEffectMatchesPermanent(g, &effects[effect], attacker) {
					lures[index]++
					p.unit = false
				}
			}
		}
		p.allowed, p.lures = append(p.allowed, allowed), append(p.lures, lures)
	}
	return p
}

func (p *blockDeclarationProblem) fixedAt(fixed []int8, blocker, attacker int) int8 {
	index := blocker*len(p.attackers) + attacker
	if index >= len(fixed) {
		return -1
	}
	return fixed[index]
}

func (p *blockDeclarationProblem) score(declarations []game.BlockDeclaration) int {
	score := 0
	counts := make(map[id.ID]int)
	for _, declaration := range declarations {
		counts[declaration.Blocking]++
		for blocker, source := range p.blockers {
			if source.ObjectID != declaration.Blocker {
				continue
			}
			for attacker, target := range p.attackers {
				if target.ObjectID == declaration.Blocking {
					score += p.lures[blocker][attacker]
				}
			}
		}
	}
	for index, attacker := range p.attackers {
		if counts[attacker.ObjectID] >= p.minimum[index] {
			score += p.weights[index]
		}
	}
	return score
}

func (p *blockDeclarationProblem) restricted(declarations []game.BlockDeclaration) bool {
	counts := make(map[id.ID]int)
	for _, declaration := range declarations {
		counts[declaration.Blocking]++
	}
	for index, attacker := range p.attackers {
		count := counts[attacker.ObjectID]
		if count > 0 && count < p.minimum[index] || count > 1 && p.oneBlocker[index] {
			return false
		}
	}
	return blockDeclarationsSatisfyAloneRestriction(p.game, appendCombatBlockers(p.game, declarations))
}

func (p *blockDeclarationProblem) complete(fixed []int8, goal int) ([]game.BlockDeclaration, int, bool) {
	if p.unit {
		declarations, score, found := p.match(fixed)
		return declarations, score, found && (goal < 0 || score >= goal)
	}
	return p.search(fixed, goal)
}

func (p *blockDeclarationProblem) maximum() (int, []game.BlockDeclaration) {
	declarations, score, found := p.complete(nil, -1)
	if !found {
		return 0, nil
	}
	return score, declarations
}

func (p *blockDeclarationProblem) match(fixed []int8) ([]game.BlockDeclaration, int, bool) {
	p.states++
	capacities := slices.Clone(p.capacities)
	counts := make([]int, len(p.attackers))
	var result []game.BlockDeclaration
	for blocker := range p.blockers {
		for attacker := range p.attackers {
			if p.fixedAt(fixed, blocker, attacker) != 1 {
				continue
			}
			if !p.allowed[blocker][attacker] || capacities[blocker] == 0 {
				return nil, 0, false
			}
			capacities[blocker]--
			counts[attacker]++
			if counts[attacker] > 1 && p.oneBlocker[attacker] {
				return nil, 0, false
			}
			result = append(result, game.BlockDeclaration{Blocker: p.blockers[blocker].ObjectID, Blocking: p.attackers[attacker].ObjectID})
		}
	}
	var slots []int
	for blocker, capacity := range capacities {
		for range min(capacity, len(p.attackers)) {
			slots = append(slots, blocker)
		}
	}
	matched := make([]int, len(slots))
	for index := range matched {
		matched[index] = -1
	}
	var augment func(int, []bool) bool
	augment = func(attacker int, seen []bool) bool {
		p.states++
		for slot, blocker := range slots {
			if seen[slot] || !p.allowed[blocker][attacker] || p.fixedAt(fixed, blocker, attacker) == 0 {
				continue
			}
			seen[slot] = true
			if matched[slot] < 0 || augment(matched[slot], seen) {
				matched[slot] = attacker
				return true
			}
		}
		return false
	}
	for attacker := range p.attackers {
		if p.weights[attacker] > 0 && counts[attacker] == 0 {
			augment(attacker, make([]bool, len(slots)))
		}
	}
	for slot, attacker := range matched {
		if attacker >= 0 {
			result = append(result, game.BlockDeclaration{Blocker: p.blockers[slots[slot]].ObjectID, Blocking: p.attackers[attacker].ObjectID})
		}
	}
	return result, p.score(result), true
}

func (p *blockDeclarationProblem) search(fixed []int8, goal int) ([]game.BlockDeclaration, int, bool) {
	best := -1
	var result []game.BlockDeclaration
	counts := make([]int, len(p.attackers))
	rootUpper := 0
	for _, weight := range p.weights {
		rootUpper += weight
	}
	for blocker := range p.blockers {
		lures := slices.Clone(p.lures[blocker])
		slices.Sort(lures)
		for index := 0; index < min(p.capacities[blocker], len(lures)); index++ {
			rootUpper += lures[len(lures)-1-index]
		}
	}
	var visit func(int, int, []game.BlockDeclaration) bool
	visit = func(blocker, lureScore int, declarations []game.BlockDeclaration) bool {
		p.states++
		upper := lureScore
		for attacker, count := range counts {
			possible := count
			for next := blocker; next < len(p.blockers); next++ {
				if p.allowed[next][attacker] && p.fixedAt(fixed, next, attacker) != 0 {
					possible++
				}
			}
			if possible >= p.minimum[attacker] {
				upper += p.weights[attacker]
			}
		}
		for next := blocker; next < len(p.blockers); next++ {
			for attacker := range p.attackers {
				if p.allowed[next][attacker] && p.fixedAt(fixed, next, attacker) != 0 {
					upper += p.lures[next][attacker]
				}
			}
		}
		if goal >= 0 && upper < goal || goal < 0 && upper <= best {
			return false
		}
		if blocker == len(p.blockers) {
			if !p.restricted(declarations) {
				return false
			}
			score := p.score(declarations)
			if score > best {
				best, result = score, slices.Clone(declarations)
			}
			return goal >= 0 && score >= goal || goal < 0 && best == rootUpper
		}
		var forced, available []int
		for attacker := range p.attackers {
			switch p.fixedAt(fixed, blocker, attacker) {
			case 1:
				if !p.allowed[blocker][attacker] {
					return false
				}
				forced = append(forced, attacker)
			case -1:
				if p.allowed[blocker][attacker] {
					available = append(available, attacker)
				}
			default:
			}
		}
		if len(forced) > p.capacities[blocker] {
			return false
		}
		for count := 0; count <= min(p.capacities[blocker]-len(forced), len(available)); count++ {
			indices := make([]game.Target, len(available))
			for index, attacker := range available {
				indices[index] = game.PermanentTarget(p.attackers[attacker].ObjectID)
			}
			for combination := range targetCombinations(indices, count) {
				chosen := slices.Clone(forced)
				for _, target := range combination {
					chosen = append(chosen, slices.IndexFunc(p.attackers, func(permanent *game.Permanent) bool { return permanent.ObjectID == target.PermanentID }))
				}
				next := slices.Clone(declarations)
				score, valid := lureScore, true
				for _, attacker := range chosen {
					counts[attacker]++
					if counts[attacker] > 1 && p.oneBlocker[attacker] {
						valid = false
					}
					next = append(next, game.BlockDeclaration{Blocker: p.blockers[blocker].ObjectID, Blocking: p.attackers[attacker].ObjectID})
					score += p.lures[blocker][attacker]
				}
				found := valid && visit(blocker+1, score, next)
				for _, attacker := range chosen {
					counts[attacker]--
				}
				if found {
					return true
				}
			}
		}
		return false
	}
	found := visit(0, 0, nil)
	if goal < 0 {
		found = best >= 0
	}
	return result, best, found
}

func blockRequirementsSatisfied(g *game.Game, player game.PlayerID, declarations []game.BlockDeclaration) bool {
	problem := newBlockDeclarationProblem(g, player)
	maximum, _ := problem.maximum()
	return problem.score(declarations) == maximum
}
