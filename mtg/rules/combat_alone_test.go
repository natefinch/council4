package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func aloneCreature(g *game.Game, controller game.PlayerID, name string, body game.StaticAbility) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:            name,
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 2}),
		Toughness:       opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{body},
	}})
}

// TestCantAttackAloneStaticBodyForbidsSoloAttack covers the "can't attack alone"
// restriction (Mogg Flunkies): the source may not be declared as the only
// attacker, but it may attack when at least one other creature also attacks.
func TestCantAttackAloneStaticBodyForbidsSoloAttack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	loner := aloneCreature(g, game.Player1, "Lonely Flunkies", game.CantAttackAloneStaticBody)
	buddy := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)

	target := game.AttackTarget{Player: game.Player2}
	solo := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: loner.ObjectID, Target: target},
	}))
	if engine.applyDeclareAttackers(g, game.Player1, solo) {
		t.Fatal("applyDeclareAttackers() accepted a solo can't-attack-alone attacker")
	}

	for _, act := range legalDeclareAttackersActions(g, game.Player1) {
		payload := mustDeclareAttackersPayload(t, act)
		if len(payload.Attackers) == 1 && payload.Attackers[0].Attacker == loner.ObjectID {
			t.Fatalf("legal attacks included the loner attacking alone: %+v", payload.Attackers)
		}
	}

	combined := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: loner.ObjectID, Target: target},
		{Attacker: buddy.ObjectID, Target: target},
	}))
	if !engine.applyDeclareAttackers(g, game.Player1, combined) {
		t.Fatal("applyDeclareAttackers() rejected the loner attacking alongside another creature")
	}
}

// TestCantBlockAloneStaticBodyForbidsSoloBlock covers the "can't block alone"
// restriction (Craven Hulk): the source may not be the only blocker in combat,
// but it may block when at least one other creature also blocks.
func TestCantBlockAloneStaticBodyForbidsSoloBlock(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attackerA := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	attackerB := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	loner := aloneCreature(g, game.Player2, "Craven Wall", game.CantBlockAloneStaticBody)
	buddy := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attackerA.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: attackerB.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)

	solo := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: loner.ObjectID, Blocking: attackerA.ObjectID},
	}))
	if engine.applyDeclareBlockers(g, game.Player2, solo) {
		t.Fatal("applyDeclareBlockers() accepted a solo can't-block-alone blocker")
	}

	for _, act := range legalDeclareBlockersActions(g, game.Player2) {
		payload := mustDeclareBlockersPayload(t, act)
		if len(payload.Blockers) == 1 && payload.Blockers[0].Blocker == loner.ObjectID {
			t.Fatalf("legal blocks included the loner blocking alone: %+v", payload.Blockers)
		}
	}

	combined := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: loner.ObjectID, Blocking: attackerA.ObjectID},
		{Blocker: buddy.ObjectID, Blocking: attackerB.ObjectID},
	}))
	if !engine.applyDeclareBlockers(g, game.Player2, combined) {
		t.Fatal("applyDeclareBlockers() rejected the loner blocking alongside another creature")
	}
}

// TestCantAttackOrBlockAloneStaticBodyForbidsSoloCombat covers the combined
// "can't attack or block alone" restriction (Loyal Pegasus): the source may
// neither be the only attacker nor the only blocker.
func TestCantAttackOrBlockAloneStaticBodyForbidsSoloCombat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	loner := aloneCreature(g, game.Player1, "Loyal Pegasus", game.CantAttackOrBlockAloneStaticBody)
	buddy := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)

	target := game.AttackTarget{Player: game.Player2}
	solo := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: loner.ObjectID, Target: target},
	}))
	if engine.applyDeclareAttackers(g, game.Player1, solo) {
		t.Fatal("applyDeclareAttackers() accepted a solo can't-attack-or-block-alone attacker")
	}
	combined := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: loner.ObjectID, Target: target},
		{Attacker: buddy.ObjectID, Target: target},
	}))
	if !engine.applyDeclareAttackers(g, game.Player1, combined) {
		t.Fatal("applyDeclareAttackers() rejected the loner attacking alongside another creature")
	}
}
