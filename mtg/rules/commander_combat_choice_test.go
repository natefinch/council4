package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/opt"
)

func TestCommanderSimulatorCombatOffersPostDamageOwnerChoice(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	commander := addCommanderPermanent(g, game.Player1)
	card, _ := g.GetCardInstance(commander.CardInstanceID)
	def := *card.Def
	def.Power, def.Toughness = opt.Val(game.PT{Value: 2}), opt.Val(game.PT{Value: 2})
	card.Def = &def
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	owner := &choiceOnlyAgent{choices: [][]int{{0}}}
	policies := simPassPolicies()
	policies[game.Player1] = owner
	policies[game.Player2] = blockingPolicy{blocker: blocker.ObjectID, attacker: commander.ObjectID}
	resolved, ok := engine.Simulator().ResolveCombatWithAttackers(g, game.Player1, action.DeclareAttackersAction{
		Attackers: []game.AttackDeclaration{{
			Attacker: commander.ObjectID, Target: game.AttackTarget{Player: game.Player2},
		}},
	}, policies)
	if !ok || owner.next != 1 ||
		!resolved.Players[game.Player1].Graveyard.Contains(commander.CardInstanceID) ||
		resolved.Players[game.Player1].CommandZone.Contains(commander.CardInstanceID) {
		t.Fatal("combat Simulator bypassed the owner's post-damage command-zone decision")
	}
	if !permanentStillPresent(g, commander.ObjectID) {
		t.Fatal("combat simulation mutated the original commander")
	}
}
