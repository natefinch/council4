package rules

import (
	"testing"

	cardm "github.com/natefinch/council4/mtg/cards/m"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

// newMBakuGame puts the real M'Baku, Jabari Chieftain onto Player1's
// battlefield next to a Player1 creature that can attack, and stages Player1's
// declare-attackers step so a test can drive the real attack trigger.
func newMBakuGame(t *testing.T) (g *game.Game, engine *Engine, attacker *game.Permanent) {
	t.Helper()
	g = game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine = NewEngine(nil)
	addCombatPermanent(g, game.Player1, cardm.MBakuJabariChieftain)
	attacker = addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1
	g.Combat = &game.CombatState{}
	return g, engine, attacker
}

// declareAttackByPlayer declares attacker as attacking defender through the real
// declare-attackers engine path on behalf of attackingPlayer, emitting the real
// EventAttackerDeclared. The shared declareAttack helper hardcodes Player1 as
// the attacking player; this variant lets a test have an opponent attack the
// controller.
func declareAttackByPlayer(t *testing.T, g *game.Game, engine *Engine, attackingPlayer game.PlayerID, attacker *game.Permanent, defender game.PlayerID) {
	t.Helper()
	g.Turn.ActivePlayer = attackingPlayer
	g.Turn.Step = game.StepDeclareAttackers
	attack := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: defender}},
	}))
	if !engine.applyDeclareAttackers(g, attackingPlayer, attack) {
		t.Fatalf("applyDeclareAttackers() rejected %v attacking %v", attacker.ObjectID, defender)
	}
}

func mustEffectiveToughness(t *testing.T, g *game.Game, permanent *game.Permanent) int {
	t.Helper()
	toughness, ok := effectiveToughness(g, permanent)
	if !ok {
		t.Fatalf("effectiveToughness(%v) has no value", permanent.ObjectID)
	}
	return toughness
}

// TestMBakuCreatureAttackingMonarchOpponentGetsBuff drives the full engine path:
// with the crown on an opponent (Player2), Player1's creature attacks that
// monarch opponent, M'Baku's trigger fires through the real enumeration, and
// resolving it gives the attacker +1/+1 and trample until end of turn.
func TestMBakuCreatureAttackingMonarchOpponentGetsBuff(t *testing.T) {
	g, engine, attacker := newMBakuGame(t)

	if got := effectivePower(g, attacker); got != 3 {
		t.Fatalf("attacker base power = %d, want 3", got)
	}
	if got := mustEffectiveToughness(t, g, attacker); got != 3 {
		t.Fatalf("attacker base toughness = %d, want 3", got)
	}

	if !setMonarch(g, game.Player2) {
		t.Fatal("setMonarch(Player2) = false")
	}
	declareAttack(t, g, engine, attacker, game.Player2)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{}}
	log := TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("M'Baku attack trigger did not fire when a creature attacked the monarch opponent")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1 (the buff)", g.Stack.Size())
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if got := effectivePower(g, attacker); got != 4 {
		t.Fatalf("attacker power after trigger = %d, want 4 (+1/+1)", got)
	}
	if got := mustEffectiveToughness(t, g, attacker); got != 4 {
		t.Fatalf("attacker toughness after trigger = %d, want 4 (+1/+1)", got)
	}
	if !hasKeyword(g, attacker, game.Trample) {
		t.Fatal("attacker did not gain trample")
	}
}

// TestMBakuCreatureAttackingNonMonarchOpponentNoBuff proves the trigger is gated
// on the attacked opponent being the monarch: when the crown is on Player3 but
// the creature attacks Player2, no ability fires and the attacker is unchanged.
func TestMBakuCreatureAttackingNonMonarchOpponentNoBuff(t *testing.T) {
	g, engine, attacker := newMBakuGame(t)

	// Player3 holds the crown; the creature attacks the non-monarch Player2.
	if !setMonarch(g, game.Player3) {
		t.Fatal("setMonarch(Player3) = false")
	}
	declareAttack(t, g, engine, attacker, game.Player2)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{}}
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("M'Baku trigger fired when the attacked opponent was not the monarch")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0 (no trigger)", g.Stack.Size())
	}
	if got := effectivePower(g, attacker); got != 3 {
		t.Fatalf("attacker power = %d, want 3 (unbuffed)", got)
	}
	if hasKeyword(g, attacker, game.Trample) {
		t.Fatal("attacker gained trample despite the attacked opponent not being the monarch")
	}
}

// TestMBakuCreatureAttackingControllerNoBuff proves the "one of your opponents"
// recipient gate excludes attacks on the controller: even when the controller
// (Player1) holds the crown, an opponent's creature attacking Player1 fires
// nothing, because Player1 is not one of Player1's opponents.
func TestMBakuCreatureAttackingControllerNoBuff(t *testing.T) {
	g, engine, _ := newMBakuGame(t)
	opponentAttacker := addCombatCreaturePermanentWithPower(g, game.Player2, 3)

	// The controller is the monarch, satisfying the intervening condition; only
	// the opponent-recipient gate should stop the trigger.
	if !setMonarch(g, game.Player1) {
		t.Fatal("setMonarch(Player1) = false")
	}
	declareAttackByPlayer(t, g, engine, game.Player2, opponentAttacker, game.Player1)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{}}
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("M'Baku trigger fired when a creature attacked the controller")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0 (no trigger)", g.Stack.Size())
	}
	if got := effectivePower(g, opponentAttacker); got != 3 {
		t.Fatalf("attacker power = %d, want 3 (unbuffed)", got)
	}
	if hasKeyword(g, opponentAttacker, game.Trample) {
		t.Fatal("attacker gained trample despite attacking the controller rather than an opponent")
	}
}

// TestMBakuBuffExpiresAtEndOfTurn proves the +1/+1 and trample last only until
// end of turn: after the buff is granted through the real path, end-of-turn
// cleanup restores the attacker to its base power/toughness and removes trample.
func TestMBakuBuffExpiresAtEndOfTurn(t *testing.T) {
	g, engine, attacker := newMBakuGame(t)

	if !setMonarch(g, game.Player2) {
		t.Fatal("setMonarch(Player2) = false")
	}
	declareAttack(t, g, engine, attacker, game.Player2)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{}}
	log := TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("M'Baku attack trigger did not fire when a creature attacked the monarch opponent")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if got := effectivePower(g, attacker); got != 4 {
		t.Fatalf("attacker power before cleanup = %d, want 4", got)
	}
	if !hasKeyword(g, attacker, game.Trample) {
		t.Fatal("attacker did not gain trample before cleanup")
	}

	expireCleanupDurations(g)

	if got := effectivePower(g, attacker); got != 3 {
		t.Fatalf("attacker power after end-of-turn cleanup = %d, want 3", got)
	}
	if got := mustEffectiveToughness(t, g, attacker); got != 3 {
		t.Fatalf("attacker toughness after end-of-turn cleanup = %d, want 3", got)
	}
	if hasKeyword(g, attacker, game.Trample) {
		t.Fatal("attacker retained trample after end-of-turn cleanup")
	}
}
