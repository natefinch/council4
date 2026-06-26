package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func manaProductionMultiplierPermanent(g *game.Game, controller game.PlayerID, name string, factor int) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{
			{RuleEffects: []game.RuleEffect{
				{Kind: game.RuleEffectManaProductionMultiplier, ManaProductionMultiplier: factor},
			}},
		},
	}})
}

func castGenericXSpell(g *game.Game, engine *Engine, x int) bool {
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Generic X Sorcery",
		ManaCost: opt.Val(cost.Mana{cost.X}),
		Types:    []types.Card{types.Sorcery},
	}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	return engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, x, nil))
}

func TestManaProductionMultiplierDoublesBasicLandPayment(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	manaProductionMultiplierPermanent(g, game.Player1, "Mana Reflection", 2)

	if !castGenericXSpell(g, engine, 2) {
		t.Fatal("casting {X} with X=2 from one doubled Forest = false, want true")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.XValue != 2 {
		t.Fatalf("stack X value = %d (ok=%v), want 2", obj.XValue, ok)
	}
}

func TestManaProductionMultiplierIsRequiredForExtraMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addBasicLandPermanent(g, game.Player1, types.Forest)

	if castGenericXSpell(g, engine, 2) {
		t.Fatal("casting {X} with X=2 from a single un-multiplied Forest = true, want false")
	}
}

func TestManaProductionMultiplierStacksMultiplicatively(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	manaProductionMultiplierPermanent(g, game.Player1, "Mana Reflection", 2)
	manaProductionMultiplierPermanent(g, game.Player1, "Nyxbloom Ancient", 3)

	if !castGenericXSpell(g, engine, 6) {
		t.Fatal("casting {X} with X=6 from one Forest under x2 and x3 multipliers = false, want true")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.XValue != 6 {
		t.Fatalf("stack X value = %d (ok=%v), want 6", obj.XValue, ok)
	}
}

func TestManaProductionMultiplierForScopesToController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	manaProductionMultiplierPermanent(g, game.Player1, "Mana Reflection", 2)
	manaProductionMultiplierPermanent(g, game.Player1, "Nyxbloom Ancient", 3)

	if got := manaProductionMultiplierFor(g, game.Player1); got != 6 {
		t.Fatalf("manaProductionMultiplierFor(Player1) = %d, want 6 (2*3)", got)
	}
	if got := manaProductionMultiplierFor(g, game.Player2); got != 1 {
		t.Fatalf("manaProductionMultiplierFor(Player2) = %d, want 1", got)
	}
}
