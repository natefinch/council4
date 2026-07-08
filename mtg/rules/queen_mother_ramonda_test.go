package rules

import (
	"testing"

	cardq "github.com/natefinch/council4/mtg/cards/q"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// newRamondaCombat puts the real Queen Mother Ramonda onto the battlefield under
// game.Player1 (its controller, i.e. "you") and a power-2 and a power-3 attacker
// under the active player game.Player2. Combat is set to the declare-attackers
// step with Player2 active. The monarch is left unset; callers set it to toggle
// Ramonda's "as long as you're the monarch" gate.
func newRamondaCombat(t *testing.T) (g *game.Game, small, big *game.Permanent) {
	t.Helper()
	g = game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, cardq.QueenMotherRamonda())
	small = addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	big = addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player2
	g.Combat = &game.CombatState{}
	return g, small, big
}

// addRamondaControllerPlaneswalker puts a planeswalker onto the battlefield under
// Ramonda's controller (game.Player1), the direct-only carve-out's target.
func addRamondaControllerPlaneswalker(g *game.Game) *game.Permanent {
	return addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:    "Ramonda Controller Planeswalker",
		Types:   []types.Card{types.Planeswalker},
		Loyalty: opt.Val(3),
	}})
}

// TestQueenMotherRamondaMonarchLowPowerCantAttackController proves ability 2
// while Ramonda's controller (game.Player1) holds the crown: a creature with
// power 2 or less can't attack the controller directly (CR 508.1 direct-only),
// but the restriction leaves other players and the controller's planeswalker
// attackable, and a creature with power 3 or more is unaffected. Every assertion
// is driven through the real declare-attackers enumeration and legality driver.
func TestQueenMotherRamondaMonarchLowPowerCantAttackController(t *testing.T) {
	g, small, big := newRamondaCombat(t)
	engine := NewEngine(nil)
	planeswalker := addRamondaControllerPlaneswalker(g)

	if !setMonarch(g, game.Player1) {
		t.Fatal("setMonarch(Player1) = false")
	}

	legal := legalDeclareAttackersActions(g, game.Player2)
	if len(legal) == 0 {
		t.Fatal("no legal declare-attackers actions")
	}

	controller := game.AttackTarget{Player: game.Player1}
	otherPlayer := game.AttackTarget{Player: game.Player3}
	controllerWalker := game.AttackTarget{Player: game.Player1, PlaneswalkerID: planeswalker.ObjectID}

	if declareAttackersActionsContainTarget(legal, small.ObjectID, controller) {
		t.Fatal("enumeration offered a power-2 creature attacking the monarch controller directly")
	}
	if !declareAttackersActionsContainTarget(legal, small.ObjectID, otherPlayer) {
		t.Fatal("enumeration omitted a power-2 creature attacking another player")
	}
	if !declareAttackersActionsContainTarget(legal, small.ObjectID, controllerWalker) {
		t.Fatal("enumeration omitted a power-2 creature attacking the controller's planeswalker (direct-only)")
	}
	if !declareAttackersActionsContainTarget(legal, big.ObjectID, controller) {
		t.Fatal("enumeration omitted a power-3 creature attacking the controller")
	}

	// The real legality driver rejects the power-2 creature attacking the
	// monarch controller directly.
	attackController := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: small.ObjectID, Target: controller},
	}))
	if engine.applyDeclareAttackers(g, game.Player2, attackController) {
		t.Fatal("applyDeclareAttackers() accepted a power-2 attack on the monarch controller")
	}
}

// TestQueenMotherRamondaMonarchLowPowerAttacksOtherPlayer proves the restriction
// is scoped to the controller only: while the controller is the monarch, a
// power-2 creature may still attack a different player. A fresh game isolates the
// mutating legality driver.
func TestQueenMotherRamondaMonarchLowPowerAttacksOtherPlayer(t *testing.T) {
	g, small, _ := newRamondaCombat(t)
	engine := NewEngine(nil)
	if !setMonarch(g, game.Player1) {
		t.Fatal("setMonarch(Player1) = false")
	}

	attackOther := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: small.ObjectID, Target: game.AttackTarget{Player: game.Player3}},
	}))
	if !engine.applyDeclareAttackers(g, game.Player2, attackOther) {
		t.Fatal("applyDeclareAttackers() rejected a power-2 attack on an unprotected player")
	}
}

// TestQueenMotherRamondaMonarchLowPowerAttacksControllerPlaneswalker proves the
// direct-only carve-out (CR 508.1) on the real legality driver: while the
// controller is the monarch, a power-2 creature may attack a planeswalker the
// controller controls even though it can't attack that player directly.
func TestQueenMotherRamondaMonarchLowPowerAttacksControllerPlaneswalker(t *testing.T) {
	g, small, _ := newRamondaCombat(t)
	engine := NewEngine(nil)
	planeswalker := addRamondaControllerPlaneswalker(g)
	if !setMonarch(g, game.Player1) {
		t.Fatal("setMonarch(Player1) = false")
	}

	attackWalker := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: small.ObjectID, Target: game.AttackTarget{Player: game.Player1, PlaneswalkerID: planeswalker.ObjectID}},
	}))
	if !engine.applyDeclareAttackers(g, game.Player2, attackWalker) {
		t.Fatal("applyDeclareAttackers() rejected a power-2 attack on the controller's planeswalker")
	}
}

// TestQueenMotherRamondaMonarchHighPowerAttacksController proves a creature with
// power 3 or more is unaffected: while the controller is the monarch, it may
// attack the controller directly. A fresh game isolates the mutating driver.
func TestQueenMotherRamondaMonarchHighPowerAttacksController(t *testing.T) {
	g, _, big := newRamondaCombat(t)
	engine := NewEngine(nil)
	if !setMonarch(g, game.Player1) {
		t.Fatal("setMonarch(Player1) = false")
	}

	attackController := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: big.ObjectID, Target: game.AttackTarget{Player: game.Player1}},
	}))
	if !engine.applyDeclareAttackers(g, game.Player2, attackController) {
		t.Fatal("applyDeclareAttackers() rejected a power-3 attack on the controller")
	}
}

// TestQueenMotherRamondaEffectivePowerDebuffBecomesRestricted proves the filter
// uses EFFECTIVE power: a power-3 creature reduced to effective power 2 by a
// -1/-1 counter can no longer attack the monarch controller.
func TestQueenMotherRamondaEffectivePowerDebuffBecomesRestricted(t *testing.T) {
	g, _, big := newRamondaCombat(t)
	engine := NewEngine(nil)
	if !setMonarch(g, game.Player1) {
		t.Fatal("setMonarch(Player1) = false")
	}

	// Before the debuff, the power-3 creature may attack the controller.
	controller := game.AttackTarget{Player: game.Player1}
	if !canAttackTarget(g, big, controller) {
		t.Fatal("power-3 creature could not attack the controller before the debuff")
	}

	// A -1/-1 counter drops its effective power to 2, bringing it under the
	// restriction.
	if !addCountersToPermanent(g, big, counter.MinusOneMinusOne, 1) {
		t.Fatal("addCountersToPermanent(-1/-1) = false")
	}
	if canAttackTarget(g, big, controller) {
		t.Fatal("a creature debuffed to effective power 2 may still attack the monarch controller")
	}
	attackController := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: big.ObjectID, Target: controller},
	}))
	if engine.applyDeclareAttackers(g, game.Player2, attackController) {
		t.Fatal("applyDeclareAttackers() accepted an attack from a creature debuffed to effective power 2")
	}
}

// TestQueenMotherRamondaNonMonarchLowPowerAttacksController proves the gate: when
// Ramonda's controller is NOT the monarch, the static's condition is unsatisfied
// so a power-2 creature may attack the controller directly. Both the enumeration
// and the real legality driver agree.
func TestQueenMotherRamondaNonMonarchLowPowerAttacksController(t *testing.T) {
	g, small, _ := newRamondaCombat(t)
	engine := NewEngine(nil)

	// No monarch: Ramonda's condition is unsatisfied, so the restriction is off.
	controller := game.AttackTarget{Player: game.Player1}
	legal := legalDeclareAttackersActions(g, game.Player2)
	if !declareAttackersActionsContainTarget(legal, small.ObjectID, controller) {
		t.Fatal("enumeration omitted a power-2 attack on the controller when no player is the monarch")
	}

	attackController := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: small.ObjectID, Target: controller},
	}))
	if !engine.applyDeclareAttackers(g, game.Player2, attackController) {
		t.Fatal("applyDeclareAttackers() rejected a power-2 attack on the controller when no player is the monarch")
	}
}
