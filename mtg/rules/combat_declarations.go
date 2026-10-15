package rules

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/opt"
)

func combatDeclarationPending(g *game.Game) bool {
	if canDeclareAttackers(g, g.Turn.ActivePlayer) {
		return true
	}
	for _, defender := range defendingPlayersInOrder(g) {
		if canDeclareBlockers(g, defender) {
			return true
		}
	}
	return false
}

func withStagedDeclaration(samples []action.Action, kind action.Kind) []action.Action {
	if len(samples) == 0 {
		return samples
	}
	stage := action.ChooseAttackers()
	if kind == action.ActionDeclareBlockers {
		stage = action.ChooseBlockers()
	}
	result := append([]action.Action(nil), samples[:len(samples)-1]...)
	result = append(result, stage, samples[len(samples)-1])
	return result
}

func (ce combatEngine) declarationLegal(g *game.Game, player game.PlayerID, declaration action.Action) bool {
	if declaration.Validate() != nil {
		return false
	}
	if attack, present := declaration.DeclareAttackersPayload(); present {
		if attack.Staged {
			return canDeclareAttackers(g, player)
		}
		return ce.attackersLegal(g, player, attack)
	}
	if block, present := declaration.DeclareBlockersPayload(); present {
		if block.Staged {
			return canDeclareBlockers(g, player)
		}
		return ce.blockersLegal(g, player, block)
	}
	return false
}

func (ce combatEngine) completeDeclaration(g *game.Game, player game.PlayerID, proposal action.Action, agents [game.NumPlayers]PlayerAgent, log *TurnLog) action.Action {
	if attack, present := proposal.DeclareAttackersPayload(); present && attack.Staged {
		return ce.chooseStagedAttackers(g, player, agents, log)
	}
	if block, present := proposal.DeclareBlockersPayload(); present && block.Staged {
		return ce.chooseStagedBlockers(g, player, agents, log)
	}
	return proposal
}

func (ce combatEngine) chooseStagedAttackers(g *game.Game, player game.PlayerID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) action.Action {
	problem := newAttackDeclarationProblem(g, player)
	goal := problem.maximum()
	fixed := make([]int, len(problem.creatures))
	for index := range fixed {
		fixed[index] = -1
	}
	for creature, domain := range problem.domains {
		var options []game.ChoiceOption
		var positions []int
		for position, choice := range domain {
			fixed[creature] = position
			if _, found := problem.complete(fixed, goal, false); !found {
				continue
			}
			option := game.ChoiceOption{Index: len(options), Label: "Do not attack"}
			if choice.attacking {
				option.Label = fmt.Sprintf("Attack with %d against %v", choice.declaration.Attacker, choice.declaration.Target)
				option.Attack = opt.Val(choice.declaration)
			}
			options, positions = append(options, option), append(positions, position)
		}
		if len(options) == 0 {
			panic("attack prefix lost its legal completion")
		}
		selected := ce.e.chooseChoice(g, agents, game.ChoiceRequest{
			Kind: game.ChoiceCombatDeclaration, Player: player, Prompt: "Choose this creature's attack",
			Subject:        permanentChoiceInfo(g, problem.creatures[creature]),
			CombatAttacker: problem.creatures[creature].ObjectID,
			Options:        options, MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{0},
		}, log)
		fixed[creature] = positions[selected[0]]
	}
	declarations, found := problem.complete(fixed, goal, false)
	if !found {
		panic("completed attack choices have no legal declaration")
	}
	return action.DeclareAttackers(declarations)
}

func (ce combatEngine) chooseStagedBlockers(g *game.Game, player game.PlayerID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) action.Action {
	problem := newBlockDeclarationProblem(g, player)
	goal, _ := problem.maximum()
	fixed := make([]int8, len(problem.blockers)*len(problem.attackers))
	for index := range fixed {
		fixed[index] = -1
	}
	for blocker, source := range problem.blockers {
		for attacker, target := range problem.attackers {
			edge := blocker*len(problem.attackers) + attacker
			if !problem.allowed[blocker][attacker] {
				fixed[edge] = 0
				continue
			}
			var options []game.ChoiceOption
			var positions []int8
			for _, position := range []int8{0, 1} {
				fixed[edge] = position
				if _, _, found := problem.complete(fixed, goal); !found {
					continue
				}
				option := game.ChoiceOption{Index: len(options), Label: "Do not block this attacker"}
				if position == 1 {
					option.Label = fmt.Sprintf("Block %d with %d", target.ObjectID, source.ObjectID)
					option.Block = opt.Val(game.BlockDeclaration{Blocker: source.ObjectID, Blocking: target.ObjectID})
				}
				options, positions = append(options, option), append(positions, position)
			}
			if len(options) == 0 {
				panic("block prefix lost its legal completion")
			}
			selected := ce.e.chooseChoice(g, agents, game.ChoiceRequest{
				Kind: game.ChoiceCombatDeclaration, Player: player, Prompt: "Choose whether to block this attacker",
				Subject:        permanentChoiceInfo(g, source),
				CombatAttacker: target.ObjectID, CombatBlocker: source.ObjectID,
				Options: options, MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{0},
			}, log)
			fixed[edge] = positions[selected[0]]
		}
	}
	declarations, _, found := problem.complete(fixed, goal)
	if !found {
		panic("completed block choices have no legal declaration")
	}
	return action.DeclareBlockers(declarations)
}
