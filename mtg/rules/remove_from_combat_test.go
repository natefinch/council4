package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestRemoveFromCombatSequenceRemovesAttackerAndUntaps(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	bystander := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	attacker.Tapped = true
	bystander.Tapped = true

	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Turn.ActivePlayer = game.Player1
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: bystander.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}}

	// "Remove target attacking creature you control from combat and untap it."
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.RemoveFromCombat{Object: game.TargetPermanentReference(0)}},
		{Primitive: game.Untap{Object: game.TargetPermanentReference(0)}},
	}, []game.Target{game.PermanentTarget(attacker.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	for _, declaration := range g.Combat.Attackers {
		if declaration.Attacker == attacker.ObjectID {
			t.Fatal("removed creature is still an attacker")
		}
	}
	if len(g.Combat.Attackers) != 1 || g.Combat.Attackers[0].Attacker != bystander.ObjectID {
		t.Fatalf("other attacker was disturbed: %+v", g.Combat.Attackers)
	}
	if attacker.Tapped {
		t.Fatal("removed creature was not untapped")
	}
	if !bystander.Tapped {
		t.Fatal("bystander should remain tapped")
	}
}
