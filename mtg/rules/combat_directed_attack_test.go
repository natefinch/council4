package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// directedMustAttackEffect builds a continuous directed forced-attack rule effect
// modeling The Brothers' War chapter II: the creatures the affected player
// controls must attack the required player (or a planeswalker or battle they
// control) each combat if able.
func directedMustAttackEffect(g *game.Game, source, affected, required game.PlayerID) game.RuleEffect {
	return game.RuleEffect{
		ID:                     g.IDGen.Next(),
		Kind:                   game.RuleEffectMustAttack,
		Controller:             source,
		AffectedSpecificPlayer: opt.Val(affected),
		RequiredAttackTarget:   opt.Val(required),
		PermanentTypes:         []types.Card{types.Creature},
	}
}

// TestDirectedMustAttackForcesAttackOnRequiredPlayer proves a directed
// forced-attack rule effect forces the affected player's creature to attack the
// required player specifically, never another opponent.
func TestDirectedMustAttackForcesAttackOnRequiredPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	forced := addCombatCreaturePermanent(g, game.Player2)
	g.Turn.ActivePlayer = game.Player2
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	g.RuleEffects = append(g.RuleEffects, directedMustAttackEffect(g, game.Player1, game.Player2, game.Player3))

	if !attackerMustAttack(g, forced) {
		t.Fatal("directed creature was not forced to attack")
	}

	legal := legalDeclareAttackersActions(g, game.Player2)
	if len(legal) == 0 {
		t.Fatal("no legal declare-attackers actions")
	}
	sawRequiredAttack := false
	for _, act := range legal {
		declarations := mustDeclareAttackersPayload(t, act)
		var forcedTarget game.AttackTarget
		attacking := false
		for _, declaration := range declarations.Attackers {
			if declaration.Attacker == forced.ObjectID {
				attacking = true
				forcedTarget = declaration.Target
			}
		}
		if !attacking {
			t.Fatalf("legal action omitted forced attacker: %+v", declarations.Attackers)
		}
		if forcedTarget.Player != game.Player3 {
			t.Fatalf("forced attacker targeted %+v, want defending player %d", forcedTarget, game.Player3)
		}
		if forcedTarget.Player == game.Player3 {
			sawRequiredAttack = true
		}
	}
	if !sawRequiredAttack {
		t.Fatal("no legal action attacked the required player")
	}
}

// TestDirectedMustAttackSatisfiedByRequiredPlayersPlaneswalker proves attacking a
// planeswalker the required player controls satisfies the directed requirement.
func TestDirectedMustAttackSatisfiedByRequiredPlayersPlaneswalker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	forced := addCombatCreaturePermanent(g, game.Player2)
	planeswalker := addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{
		Name:    "Test Planeswalker",
		Types:   []types.Card{types.Planeswalker},
		Loyalty: opt.Val(3),
	}})
	g.Turn.ActivePlayer = game.Player2
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	g.RuleEffects = append(g.RuleEffects, directedMustAttackEffect(g, game.Player1, game.Player2, game.Player3))

	want := game.AttackTarget{Player: game.Player3, PlaneswalkerID: planeswalker.ObjectID}
	legal := legalDeclareAttackersActions(g, game.Player2)
	if !declareAttackersActionsContainTarget(legal, forced.ObjectID, want) {
		t.Fatalf("legal actions = %+v, want planeswalker target %v", legal, want)
	}
	engine := NewEngine(nil)
	declare := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: forced.ObjectID, Target: want},
	}))
	if !engine.applyDeclareAttackers(g, game.Player2, declare) {
		t.Fatal("applyDeclareAttackers() rejected attack on required player's planeswalker")
	}
}

// TestDirectedMustAttackRejectsWrongDefender proves a directed creature may not
// attack a defender other than the required player while the required player is a
// reachable target.
func TestDirectedMustAttackRejectsWrongDefender(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	forced := addCombatCreaturePermanent(g, game.Player2)
	g.Turn.ActivePlayer = game.Player2
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	g.RuleEffects = append(g.RuleEffects, directedMustAttackEffect(g, game.Player1, game.Player2, game.Player3))

	engine := NewEngine(nil)
	declare := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: forced.ObjectID, Target: game.AttackTarget{Player: game.Player1}},
	}))
	if engine.applyDeclareAttackers(g, game.Player2, declare) {
		t.Fatal("applyDeclareAttackers() accepted attack on the wrong defender")
	}
}

// TestDirectedMustAttackNotForcedWhenRequiredPlayerEliminated proves the directed
// requirement does not bind when the required player is no longer a legal target
// ("... each combat if able"), leaving the creature free not to attack and free
// to attack any other defender.
func TestDirectedMustAttackNotForcedWhenRequiredPlayerEliminated(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	forced := addCombatCreaturePermanent(g, game.Player2)
	g.Players[game.Player3].Eliminated = true
	g.Turn.ActivePlayer = game.Player2
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	g.RuleEffects = append(g.RuleEffects, directedMustAttackEffect(g, game.Player1, game.Player2, game.Player3))

	engine := NewEngine(nil)
	declareElsewhere := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: forced.ObjectID, Target: game.AttackTarget{Player: game.Player1}},
	}))
	if !engine.applyDeclareAttackers(g, game.Player2, declareElsewhere) {
		t.Fatal("applyDeclareAttackers() rejected free attack when required player eliminated")
	}

	g.Combat = &game.CombatState{}
	declareNone := mustDeclareAttackersPayload(t, action.DeclareAttackers(nil))
	if !engine.applyDeclareAttackers(g, game.Player2, declareNone) {
		t.Fatal("applyDeclareAttackers() forced an attack when required player eliminated")
	}
}
