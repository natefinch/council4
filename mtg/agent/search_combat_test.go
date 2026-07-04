package agent

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

// combatWorld builds a game at the declare-attackers step with Player1 active,
// so searchAttackers can evaluate candidate attacks through real combat.
func combatWorld() *game.Game {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	return g
}

func attackCandidates(attacker *game.Permanent, defender game.PlayerID) []action.Action {
	return []action.Action{
		action.DeclareAttackers([]game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: defender}},
		}),
		action.DeclareAttackers(nil), // the no-attack option
	}
}

func TestSearchAttackersSwingsForValueIntoOpenBoard(t *testing.T) {
	e := searchTestEngine()
	g := combatWorld()
	attacker := addObservedPermanent(g, game.Player1, creatureCardDef("Knight", 4, 4))
	// Player2 has no blockers, so the attack connects for free value.
	searcher := Searcher{Rollout: GenericStrategy{}}

	chosen := searcher.searchAttackers(e.Simulator(), g, game.Player1, attackCandidates(attacker, game.Player2))
	payload, ok := chosen.DeclareAttackersPayload()
	if !ok || len(payload.Attackers) == 0 {
		t.Fatalf("searcher chose %v, want to attack into an open board", chosen)
	}
}

func TestSearchAttackersDeclinesSuicidalAttack(t *testing.T) {
	e := searchTestEngine()
	g := combatWorld()
	attacker := addObservedPermanent(g, game.Player1, creatureCardDef("Squire", 2, 2))
	// Player2 has a 5/5 that the rollout policy will block with, killing the 2/2
	// for free, so attacking loses a creature for nothing.
	addObservedPermanent(g, game.Player2, creatureCardDef("Ogre", 5, 5))
	searcher := Searcher{Rollout: GenericStrategy{}}

	chosen := searcher.searchAttackers(e.Simulator(), g, game.Player1, attackCandidates(attacker, game.Player2))
	if payload, ok := chosen.DeclareAttackersPayload(); ok && len(payload.Attackers) > 0 {
		t.Fatalf("searcher attacked into a lethal blocker (%v); want to hold back", chosen)
	}
}

func TestChooseActionBySearchRoutesAttackDeclarations(t *testing.T) {
	// isAttackDeclaration must classify declare-attackers legal sets so combat
	// search is used rather than the priority-action path.
	if !isAttackDeclaration([]action.Action{action.DeclareAttackers(nil)}) {
		t.Fatal("isAttackDeclaration did not recognize a declare-attackers action")
	}
	if isAttackDeclaration([]action.Action{action.Pass()}) {
		t.Fatal("isAttackDeclaration wrongly classified a pass as an attack declaration")
	}
}
