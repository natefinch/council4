package rules

import (
	"cmp"
	"slices"

	"github.com/natefinch/council4/mtg/game"
)

type attackChoice struct {
	declaration game.AttackDeclaration
	attacking   bool
	score       int
	taxed       bool
}

type attackDeclarationProblem struct {
	game      *game.Game
	player    game.PlayerID
	creatures []*game.Permanent
	domains   [][]attackChoice
	states    int
}

func newAttackDeclarationProblem(g *game.Game, player game.PlayerID) *attackDeclarationProblem {
	g.BeginStaticSourceFrame()
	defer g.EndStaticSourceFrame()
	problem := &attackDeclarationProblem{game: g, player: player, creatures: eligibleAttackers(g, player)}
	targets := legalAttackTargets(g, player)
	effects := activeRuleEffects(g)
	for _, creature := range problem.creatures {
		domain := []attackChoice{{}}
		for _, target := range targets {
			if !canAttackTarget(g, creature, target) {
				continue
			}
			choice := attackChoice{attacking: true, declaration: game.AttackDeclaration{Attacker: creature.ObjectID, Target: target}}
			for index := range effects {
				effect := &effects[index]
				if effect.Kind == game.RuleEffectMustAttack && ruleEffectMatchesPermanent(g, effect, creature) &&
					(!effect.RequiredAttackTarget.Exists || effect.RequiredAttackTarget.Val == target.Player) {
					choice.score++
				}
			}
			for goader := range game.PlayerID(game.NumPlayers) {
				if !wasGoadedByNow(g, creature, goader) {
					continue
				}
				choice.score++
				if target.IsPlayerAttack() && target.Player != goader {
					choice.score++
				}
			}
			_, choice.taxed = (combatEngine{}).attackTaxCost(g, []game.AttackDeclaration{choice.declaration})
			domain = append(domain, choice)
		}
		slices.SortStableFunc(domain, func(a, b attackChoice) int {
			return cmp.Compare(b.score, a.score)
		})
		problem.domains = append(problem.domains, domain)
	}
	return problem
}

func (p *attackDeclarationProblem) score(declarations []game.AttackDeclaration) int {
	score := 0
	for _, declaration := range declarations {
		for index, creature := range p.creatures {
			if creature.ObjectID != declaration.Attacker {
				continue
			}
			for _, choice := range p.domains[index] {
				if choice.attacking && choice.declaration == declaration {
					score += choice.score
					break
				}
			}
			break
		}
	}
	return score
}

func (p *attackDeclarationProblem) upperBound(fixed []int, start int, untaxed bool) int {
	upper := 0
	for index := start; index < len(p.domains); index++ {
		best := 0
		for position, choice := range p.domains[index] {
			if len(fixed) > index && fixed[index] >= 0 && fixed[index] != position || untaxed && choice.taxed {
				continue
			}
			best = max(best, choice.score)
		}
		upper += best
	}
	return upper
}

func (p *attackDeclarationProblem) maximum() int {
	for goal := p.upperBound(nil, 0, true); goal > 0; goal-- {
		if _, found := p.complete(nil, goal, true); found {
			return goal
		}
	}
	return 0
}

// Voluntary taxed attacks authorize only their chosen costs, not payment for
// other creatures or defenders. They cannot excuse freely satisfiable duties.
func (p *attackDeclarationProblem) withPaidChoices(declarations []game.AttackDeclaration) *attackDeclarationProblem {
	enabled := *p
	enabled.domains = slices.Clone(p.domains)
	for creature, domain := range p.domains {
		for position, choice := range domain {
			if choice.taxed && slices.Contains(declarations, choice.declaration) {
				enabled.domains[creature] = slices.Clone(domain)
				enabled.domains[creature][position].taxed = false
				break
			}
		}
	}
	return &enabled
}

func (p *attackDeclarationProblem) complete(fixed []int, goal int, untaxed bool) ([]game.AttackDeclaration, bool) {
	eligible := permanentMapByObjectID(p.creatures)
	var result []game.AttackDeclaration
	var visit func(int, int, []game.AttackDeclaration) bool
	visit = func(index, score int, declarations []game.AttackDeclaration) bool {
		p.states++
		if score+p.upperBound(fixed, index, untaxed) < goal {
			return false
		}
		if index == len(p.domains) {
			if score < goal || !attackDeclarationsSatisfyAloneRestriction(p.game, declarations, eligible) ||
				!(combatEngine{}).canPayAttackTax(p.game, p.player, declarations) {
				return false
			}
			if !untaxed && score < p.withPaidChoices(declarations).maximum() {
				return false
			}
			result = slices.Clone(declarations)
			return true
		}
		for position, choice := range p.domains[index] {
			if len(fixed) > index && fixed[index] >= 0 && fixed[index] != position || untaxed && choice.taxed {
				continue
			}
			next := declarations
			if choice.attacking {
				next = append(slices.Clone(declarations), choice.declaration)
			}
			if visit(index+1, score+choice.score, next) {
				return true
			}
		}
		return false
	}
	found := visit(0, 0, nil)
	return result, found
}

func attackRequirementsSatisfied(g *game.Game, player game.PlayerID, declarations []game.AttackDeclaration) bool {
	problem := newAttackDeclarationProblem(g, player)
	score := problem.score(declarations)
	problem = problem.withPaidChoices(declarations)
	if score >= problem.upperBound(nil, 0, true) {
		return true
	}
	return score >= problem.maximum()
}
