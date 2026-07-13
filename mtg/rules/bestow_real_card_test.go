package rules

import (
	"testing"

	cards "github.com/natefinch/council4/mtg/cards/n"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestNyxbornRollickerBestowGeneratedCard exercises the real, compiler-generated
// Nyxborn Rollicker ("Bestow {1}{R}\nEnchanted creature gets +1/+1.") end to end:
// cast bestowed it becomes an Aura that attaches to a creature, stops being a
// creature, and grants the enchanted creature +1/+1. This proves the generated
// BestowStaticAbility flows through cast, resolution, attachment, and the layer
// system exactly like the hand-crafted reference card.
func TestNyxbornRollickerBestowGeneratedCard(t *testing.T) {
	g, engine := setupBestowMain(t)
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	spellID := addCardToHand(g, game.Player1, cards.NyxbornRollicker())
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)

	targets := []game.Target{game.PermanentTarget(target.ObjectID)}
	if !engine.applyAction(g, game.Player1, action.CastBestowSpell(spellID, targets, 0, nil)) {
		t.Fatal("bestowed cast of Nyxborn Rollicker failed")
	}
	if obj, ok := g.Stack.Peek(); !ok || !obj.Bestowed {
		t.Fatalf("stack object = %#v, want Bestowed spell", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	aura, ok := findPermanentByCardID(g, spellID)
	if !ok {
		t.Fatal("Nyxborn Rollicker did not enter the battlefield")
	}
	if !aura.Bestowed {
		t.Fatal("Nyxborn Rollicker not marked Bestowed")
	}
	if !aura.AttachedTo.Exists || aura.AttachedTo.Val != target.ObjectID {
		t.Fatalf("Nyxborn Rollicker attached to = %v, want creature %v", aura.AttachedTo, target.ObjectID)
	}
	if permanentHasType(g, aura, types.Creature) {
		t.Fatal("bestowed Nyxborn Rollicker is still a creature, want Aura only")
	}
	if !permanentHasSubtype(g, aura, types.Aura) {
		t.Fatal("bestowed Nyxborn Rollicker is not an Aura")
	}
	if got := effectivePower(g, target); got != 3 {
		t.Fatalf("enchanted creature power = %d, want 3", got)
	}
	if got, _ := effectiveToughness(g, target); got != 3 {
		t.Fatalf("enchanted creature toughness = %d, want 3", got)
	}
}

// TestNyxbornRollickerNormalCastIsCreature proves the same generated card cast
// for its {R} mana cost enters as an ordinary creature that takes no target and
// is not an Aura.
func TestNyxbornRollickerNormalCastIsCreature(t *testing.T) {
	g, engine := setupBestowMain(t)
	spellID := addCardToHand(g, game.Player1, cards.NyxbornRollicker())
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("normal cast of Nyxborn Rollicker failed")
	}
	if obj, ok := g.Stack.Peek(); !ok || obj.Bestowed {
		t.Fatalf("stack object = %#v, want non-bestowed creature spell", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	permanent, ok := findPermanentByCardID(g, spellID)
	if !ok {
		t.Fatal("Nyxborn Rollicker did not enter the battlefield")
	}
	if permanent.Bestowed || permanent.AttachedTo.Exists {
		t.Fatalf("normally cast Nyxborn Rollicker is bestowed/attached: %#v", permanent)
	}
	if !permanentHasType(g, permanent, types.Creature) {
		t.Fatal("normally cast Nyxborn Rollicker is not a creature")
	}
	if isAuraPermanent(g, permanent) {
		t.Fatal("normally cast Nyxborn Rollicker is an Aura")
	}
}
