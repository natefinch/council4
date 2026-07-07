package rules

import (
	"slices"
	"testing"

	cardf "github.com/natefinch/council4/mtg/cards/f"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// attachFealtyToCreature puts the real Fealty to the Realm card onto the
// battlefield under game.Player1's control (the Aura's controller, i.e. "you")
// and attaches it to a freshly created creature that normally belongs to
// game.Player2. It returns the enchanted creature and the Aura permanent.
func attachFealtyToCreature(t *testing.T) (g *game.Game, creature, aura *game.Permanent) {
	t.Helper()
	g = game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature = addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	aura = addCombatPermanent(g, game.Player1, cardf.FealtyToTheRealm)
	if !attachPermanent(g, aura, creature) {
		t.Fatal("attachPermanent(Fealty to the Realm, creature) = false")
	}
	return g, creature, aura
}

// TestFealtyToTheRealmMonarchControlsEnchantedCreature proves the dynamic
// control-transfer static ("The monarch controls enchanted creature."): control
// follows the crown as it moves between players, and reverts to the creature's
// normal controller whenever no player is the monarch.
func TestFealtyToTheRealmMonarchControlsEnchantedCreature(t *testing.T) {
	g, creature, _ := attachFealtyToCreature(t)

	// No monarch: the enchanted creature keeps its normal controller.
	if got := effectiveController(g, creature); got != game.Player2 {
		t.Fatalf("effectiveController with no monarch = %v, want Player2", got)
	}

	// The monarch controls the creature: control moves to the crown holder.
	if !setMonarch(g, game.Player1) {
		t.Fatal("setMonarch(Player1) = false")
	}
	if got := effectiveController(g, creature); got != game.Player1 {
		t.Fatalf("effectiveController with monarch Player1 = %v, want Player1", got)
	}

	// The crown moves: control follows to the new monarch.
	if !setMonarch(g, game.Player3) {
		t.Fatal("setMonarch(Player3) = false")
	}
	if got := effectiveController(g, creature); got != game.Player3 {
		t.Fatalf("effectiveController with monarch Player3 = %v, want Player3", got)
	}

	// No monarch again: control reverts to the creature's normal controller.
	for i := range g.Players {
		g.Players[i].IsMonarch = false
	}
	if got := effectiveController(g, creature); got != game.Player2 {
		t.Fatalf("effectiveController after monarch cleared = %v, want Player2", got)
	}
}

// TestFealtyToTheRealmEliminatedMonarchDoesNotControl proves the enchanted
// creature is not controlled by a monarch who has left the game. IsMonarch is
// not cleared on elimination (only setMonarch unsets it), so an eliminated
// crown holder still carries the flag; the continuous control effect must ignore
// them (CR 800.4a) and leave the creature with its normal controller.
func TestFealtyToTheRealmEliminatedMonarchDoesNotControl(t *testing.T) {
	g, creature, _ := attachFealtyToCreature(t)

	if !setMonarch(g, game.Player3) {
		t.Fatal("setMonarch(Player3) = false")
	}
	if got := effectiveController(g, creature); got != game.Player3 {
		t.Fatalf("effectiveController with living monarch Player3 = %v, want Player3", got)
	}

	// Player3 leaves the game while still holding the crown (a non-combat loss
	// leaves IsMonarch set).
	g.Players[game.Player3].Eliminated = true
	if got := effectiveController(g, creature); got != game.Player2 {
		t.Fatalf("effectiveController with eliminated monarch = %v, want the normal controller Player2", got)
	}
	// The full continuous-value path (applyContinuousEffect) must agree, so the
	// second NewControllerIsMonarch site is covered too.
	if got := effectivePermanentValues(g, creature).controller; got != game.Player2 {
		t.Fatalf("effectivePermanentValues controller with eliminated monarch = %v, want Player2", got)
	}
}

// TestFealtyToTheRealmEnchantedCreatureMustAttackEachCombat proves the
// "attacks each combat if able" half of ability 3 through the real
// declare-attackers driver: every legal declaration forces the enchanted
// creature in, and declining to attack is rejected.
func TestFealtyToTheRealmEnchantedCreatureMustAttackEachCombat(t *testing.T) {
	g, creature, _ := attachFealtyToCreature(t)
	engine := NewEngine(nil)

	// The monarch (Player3) controls the enchanted creature and is the active
	// player declaring attackers.
	if !setMonarch(g, game.Player3) {
		t.Fatal("setMonarch(Player3) = false")
	}
	if got := effectiveController(g, creature); got != game.Player3 {
		t.Fatalf("effectiveController = %v, want monarch Player3", got)
	}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player3
	g.Combat = &game.CombatState{}

	legal := legalDeclareAttackersActions(g, game.Player3)
	if len(legal) == 0 {
		t.Fatal("no legal declare-attackers actions for the monarch")
	}
	for _, act := range legal {
		declarations := mustDeclareAttackersPayload(t, act)
		if !slices.ContainsFunc(declarations.Attackers, func(d game.AttackDeclaration) bool {
			return d.Attacker == creature.ObjectID
		}) {
			t.Fatalf("legal action omitted the required enchanted attacker: %+v", declarations.Attackers)
		}
	}

	// The real driver rejects declining to attack with the forced creature.
	if engine.applyDeclareAttackers(g, game.Player3, mustDeclareAttackersPayload(t, action.DeclareAttackers(nil))) {
		t.Fatal("applyDeclareAttackers() accepted an empty declaration despite must-attack")
	}

	// A legal attack that satisfies must-attack is accepted.
	legalAttack := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: creature.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}))
	if !engine.applyDeclareAttackers(g, game.Player3, legalAttack) {
		t.Fatal("applyDeclareAttackers() rejected a legal must-attack declaration")
	}
}

// TestFealtyToTheRealmEnchantedCreatureCantAttackAuraController proves the
// "can't attack you" half of ability 3: "you" is the Aura's controller
// (game.Player1), even though the monarch (game.Player3) now controls the
// creature. The restriction is direct-only (CR 508.1): the creature can't
// attack the Aura controller as a player, but can attack other players and can
// still attack a planeswalker the Aura controller controls. Every assertion is
// driven through the real declare-attackers enumeration and legality driver.
func TestFealtyToTheRealmEnchantedCreatureCantAttackAuraController(t *testing.T) {
	g, creature, _ := attachFealtyToCreature(t)
	engine := NewEngine(nil)
	planeswalker := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:    "Aura Controller Planeswalker",
		Types:   []types.Card{types.Planeswalker},
		Loyalty: opt.Val(3),
	}})

	// The monarch (Player3) controls the creature and declares attackers.
	if !setMonarch(g, game.Player3) {
		t.Fatal("setMonarch(Player3) = false")
	}
	if got := effectiveController(g, creature); got != game.Player3 {
		t.Fatalf("effectiveController = %v, want monarch Player3", got)
	}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player3
	g.Combat = &game.CombatState{}

	legal := legalDeclareAttackersActions(g, game.Player3)

	auraController := game.AttackTarget{Player: game.Player1}
	otherPlayer := game.AttackTarget{Player: game.Player2}
	auraControllerWalker := game.AttackTarget{Player: game.Player1, PlaneswalkerID: planeswalker.ObjectID}

	if declareAttackersActionsContainTarget(legal, creature.ObjectID, auraController) {
		t.Fatal("enumeration offered attacking the Aura controller directly")
	}
	if !declareAttackersActionsContainTarget(legal, creature.ObjectID, otherPlayer) {
		t.Fatal("enumeration omitted attacking a player the creature may attack")
	}
	if !declareAttackersActionsContainTarget(legal, creature.ObjectID, auraControllerWalker) {
		t.Fatal("enumeration omitted attacking the Aura controller's planeswalker (direct-only)")
	}

	// The real legality driver rejects attacking the Aura controller directly.
	attackYou := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: creature.ObjectID, Target: auraController},
	}))
	if engine.applyDeclareAttackers(g, game.Player3, attackYou) {
		t.Fatal("applyDeclareAttackers() accepted an attack on the Aura controller")
	}

	// Attacking another player is legal.
	attackOther := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: creature.ObjectID, Target: otherPlayer},
	}))
	if !engine.applyDeclareAttackers(g, game.Player3, attackOther) {
		t.Fatal("applyDeclareAttackers() rejected an attack on an unprotected player")
	}
}

// TestFealtyToTheRealmEnchantedCreatureCanAttackAuraControllerPlaneswalker
// isolates the direct-only carve-out on the real legality driver in a fresh
// game (applyDeclareAttackers mutates combat state): the enchanted creature can
// attack a planeswalker the Aura controller controls even though it can't
// attack that player directly.
func TestFealtyToTheRealmEnchantedCreatureCanAttackAuraControllerPlaneswalker(t *testing.T) {
	g, creature, _ := attachFealtyToCreature(t)
	engine := NewEngine(nil)
	planeswalker := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:    "Aura Controller Planeswalker",
		Types:   []types.Card{types.Planeswalker},
		Loyalty: opt.Val(3),
	}})

	if !setMonarch(g, game.Player3) {
		t.Fatal("setMonarch(Player3) = false")
	}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player3
	g.Combat = &game.CombatState{}

	attackWalker := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: creature.ObjectID, Target: game.AttackTarget{Player: game.Player1, PlaneswalkerID: planeswalker.ObjectID}},
	}))
	if !engine.applyDeclareAttackers(g, game.Player3, attackWalker) {
		t.Fatal("applyDeclareAttackers() rejected an attack on the Aura controller's planeswalker")
	}
}
