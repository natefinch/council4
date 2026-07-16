package payment

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
)

func TestExactManaIncreaseComposesWithGenericModifiers(t *testing.T) {
	t.Parallel()
	base := cost.Mana{cost.O(2), cost.W}
	got := applyManaCostModifiers(&base, []game.CostModifier{
		{Kind: game.CostModifierSpell, ManaIncrease: cost.Mana{cost.O(1), cost.W}},
		{Kind: game.CostModifierSpell, GenericIncrease: 1},
		{Kind: game.CostModifierSpell, GenericReduction: 2},
	})
	want := cost.Mana{cost.O(2), cost.W, cost.W}
	if got == nil || !slices.Equal(*got, want) {
		t.Fatalf("modified cost = %#v, want %#v", got, want)
	}
}

func TestExactManaIncreaseAppliesToFreeAlternativeCost(t *testing.T) {
	t.Parallel()
	free := cost.Mana{}
	got := applyManaCostModifiers(&free, []game.CostModifier{
		{Kind: game.CostModifierSpell, ManaIncrease: cost.Mana{cost.O(1), cost.W}},
		{Kind: game.CostModifierSpell, GenericReduction: 1},
	})
	if got == nil || !slices.Equal(*got, cost.Mana{cost.W}) {
		t.Fatalf("modified free cost = %#v, want {W}", got)
	}
}
