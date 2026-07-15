package rules

import (
	"testing"

	cards "github.com/natefinch/council4/mtg/cards/l"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// leylineAxeBearPermanent puts a plain 2/2 creature onto the battlefield under
// controller, to serve as an equip target.
func leylineAxeBearPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Runeclaw Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
}

// TestLeylineAxeEquipsAndBuffsEquippedCreature drives the real card's "Equip
// {3}" activated ability through the stack and proves the equipped creature
// gains +1/+1, double strike, and trample only once the equipment is attached.
func TestLeylineAxeEquipsAndBuffsEquippedCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	axe := addCombatPermanent(g, game.Player1, cards.LeylineAxe())
	bear := leylineAxeBearPermanent(g, game.Player1)
	for range 3 {
		addBasicLandPermanent(g, game.Player1, types.Forest)
	}
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	// Before attachment the creature is an unmodified 2/2 with no granted keywords.
	if got := effectivePower(g, bear); got != 2 {
		t.Fatalf("bear power before equip = %d, want 2", got)
	}
	if hasKeyword(g, bear, game.DoubleStrike) || hasKeyword(g, bear, game.Trample) {
		t.Fatal("bear had granted keywords before the Axe was attached")
	}

	act := action.ActivateAbility(axe.ObjectID, 0, []game.Target{game.PermanentTarget(bear.ObjectID)}, 0)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("equip activation was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(equip) = false, want true")
	}
	// Equip goes on the stack; the buff applies only after it resolves and attaches.
	if axe.AttachedTo.Exists {
		t.Fatal("Axe attached before its equip ability resolved")
	}
	engine.resolveTopOfStack(g, nil)

	if !axe.AttachedTo.Exists || axe.AttachedTo.Val != bear.ObjectID {
		t.Fatalf("Axe attached to = %v, want %v", axe.AttachedTo, bear.ObjectID)
	}
	if !permanentIDsContain(bear.Attachments, axe.ObjectID) {
		t.Fatal("equipped creature does not reference the Axe")
	}
	if got := effectivePower(g, bear); got != 3 {
		t.Fatalf("equipped bear power = %d, want 3", got)
	}
	toughness, ok := effectiveToughness(g, bear)
	if !ok || toughness != 3 {
		t.Fatalf("equipped bear toughness = %d (ok=%v), want 3", toughness, ok)
	}
	if !hasKeyword(g, bear, game.DoubleStrike) {
		t.Fatal("equipped bear did not gain double strike")
	}
	if !hasKeyword(g, bear, game.Trample) {
		t.Fatal("equipped bear did not gain trample")
	}
}

// TestLeylineAxeBuffMovesWithTheEquipment proves the +1/+1 and keyword grants
// track the equipment: a second creature that the Axe re-equips gains them, and
// the first creature loses them.
func TestLeylineAxeBuffMovesWithTheEquipment(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	axe := addCombatPermanent(g, game.Player1, cards.LeylineAxe())
	first := leylineAxeBearPermanent(g, game.Player1)
	second := leylineAxeBearPermanent(g, game.Player1)
	for range 6 {
		addBasicLandPermanent(g, game.Player1, types.Forest)
	}
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	equip := func(target *game.Permanent) {
		t.Helper()
		act := action.ActivateAbility(axe.ObjectID, 0, []game.Target{game.PermanentTarget(target.ObjectID)}, 0)
		if !engine.applyAction(g, game.Player1, act) {
			t.Fatalf("applyAction(equip %v) = false, want true", target.ObjectID)
		}
		engine.resolveTopOfStack(g, nil)
	}

	equip(first)
	if got := effectivePower(g, first); got != 3 {
		t.Fatalf("first creature power while equipped = %d, want 3", got)
	}

	equip(second)
	if got := effectivePower(g, second); got != 3 {
		t.Fatalf("second creature power after re-equip = %d, want 3", got)
	}
	if got := effectivePower(g, first); got != 2 {
		t.Fatalf("first creature power after Axe moved away = %d, want 2", got)
	}
	if hasKeyword(g, first, game.DoubleStrike) || hasKeyword(g, first, game.Trample) {
		t.Fatal("first creature kept granted keywords after the Axe moved away")
	}
}
