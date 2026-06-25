package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/opt"
)

// mangaraAttackingControllerPermanent installs a "Whenever an opponent attacks
// with creatures, if two or more of those creatures are attacking you and/or
// planeswalkers you control, draw a card" trigger controlled by the given
// player and returns the permanent.
func mangaraAttackingControllerPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	permanent := addTriggeredPermanent(g, controller, &game.TriggerPattern{
		Event:      game.EventAttackerDeclared,
		Controller: game.TriggerControllerOpponent,
		OneOrMore:  true,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	card, ok := g.GetCardInstance(permanent.CardInstanceID)
	if !ok {
		panic("triggered permanent card instance not found")
	}
	card.Def.TriggeredAbilities[0].Trigger.InterveningCondition = opt.Val(game.Condition{
		Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateAttackersAttackingController, Op: compare.GreaterOrEqual, Value: 2}},
	})
	return permanent
}

// TestAttackingControllerTriggerFiresWhenTwoAttackController verifies the Mangara
// intervening-if fires when two or more of an opponent's attackers are attacking
// the trigger's controller.
func TestAttackingControllerTriggerFiresWhenTwoAttackController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	mangaraAttackingControllerPermanent(g, game.Player1)
	first := addCombatCreaturePermanent(g, game.Player2)
	second := addCombatCreaturePermanent(g, game.Player2)
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: first.ObjectID, Target: game.AttackTarget{Player: game.Player1}},
		{Attacker: second.ObjectID, Target: game.AttackTarget{Player: game.Player1}},
	}}
	batchID := g.IDGen.Next()
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player2, PermanentID: first.ObjectID, SimultaneousID: batchID})
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player2, PermanentID: second.ObjectID, SimultaneousID: batchID})

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("attacking-controller trigger was not put on stack when two attackers attacked the controller")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want a single coalesced trigger", got)
	}
}

// TestAttackingControllerTriggerDoesNotFireBelowThreshold verifies the trigger
// does not fire when only one of the opponent's attackers is attacking the
// controller, even though several creatures attack overall.
func TestAttackingControllerTriggerDoesNotFireBelowThreshold(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	mangaraAttackingControllerPermanent(g, game.Player1)
	atController := addCombatCreaturePermanent(g, game.Player2)
	atOther := addCombatCreaturePermanent(g, game.Player2)
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: atController.ObjectID, Target: game.AttackTarget{Player: game.Player1}},
		{Attacker: atOther.ObjectID, Target: game.AttackTarget{Player: game.Player3}},
	}}
	batchID := g.IDGen.Next()
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player2, PermanentID: atController.ObjectID, SimultaneousID: batchID})
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player2, PermanentID: atOther.ObjectID, SimultaneousID: batchID})

	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("attacking-controller trigger fired with only one attacker attacking the controller")
	}
	if got := g.Stack.Size(); got != 0 {
		t.Fatalf("stack size = %d, want no trigger", got)
	}
}

// TestAttackingControllerTriggerDoesNotFireAttackingOther verifies the trigger
// does not fire when the opponent's attackers all attack a different player.
func TestAttackingControllerTriggerDoesNotFireAttackingOther(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	mangaraAttackingControllerPermanent(g, game.Player1)
	first := addCombatCreaturePermanent(g, game.Player2)
	second := addCombatCreaturePermanent(g, game.Player2)
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: first.ObjectID, Target: game.AttackTarget{Player: game.Player3}},
		{Attacker: second.ObjectID, Target: game.AttackTarget{Player: game.Player3}},
	}}
	batchID := g.IDGen.Next()
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player2, PermanentID: first.ObjectID, SimultaneousID: batchID})
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player2, PermanentID: second.ObjectID, SimultaneousID: batchID})

	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("attacking-controller trigger fired when attackers attacked a different player")
	}
	if got := g.Stack.Size(); got != 0 {
		t.Fatalf("stack size = %d, want no trigger", got)
	}
}

// TestAttackersAttackingPlayerCount checks that the helper counts attackers
// attacking the player directly or one of their planeswalkers, and excludes
// battle attacks.
func TestAttackersAttackingPlayerCount(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Target: game.AttackTarget{Player: game.Player1}},
		{Target: game.AttackTarget{Player: game.Player1, PlaneswalkerID: id.ID(42)}},
		{Target: game.AttackTarget{Player: game.Player1, BattleID: id.ID(43)}},
		{Target: game.AttackTarget{Player: game.Player2}},
	}}
	if got := attackersAttackingPlayerCount(g, game.Player1); got != 2 {
		t.Fatalf("attackersAttackingPlayerCount(Player1) = %d, want 2 (direct + planeswalker, excluding battle and other player)", got)
	}
	if got := attackersAttackingPlayerCount(g, game.Player2); got != 1 {
		t.Fatalf("attackersAttackingPlayerCount(Player2) = %d, want 1", got)
	}
}
