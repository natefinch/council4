package rules

import (
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/opt"
)

// regColoredSymbolCount counts the colored mana symbols in a cost so tests can
// assert that generic-only cost modifiers never alter colored requirements.
func regColoredSymbolCount(manaCost *cost.Mana) int {
	if manaCost == nil {
		return 0
	}
	count := 0
	for _, symbol := range *manaCost {
		if symbol.Kind == cost.ColoredSymbol {
			count++
		}
	}
	return count
}

// regGreenGreenThree is a representative {3}{G}{G} starting cost.
func regGreenGreenThree() cost.Mana {
	return cost.Mana{cost.O(3), cost.G, cost.G}
}

// TestRegCostModifiersStackIncreasesAndReductions checks that several generic
// increases and reductions are summed, leaving colored requirements untouched.
func TestRegCostModifiersStackIncreasesAndReductions(t *testing.T) {
	t.Parallel()
	base := regGreenGreenThree()
	modifiers := []game.CostModifier{
		{GenericIncrease: 2},
		{GenericIncrease: 1},
		{GenericReduction: 1},
	}

	result := applyAbilityCostModifiers(&base, modifiers)

	// 3 + 2 + 1 - 1 = 5 generic; the two {G} symbols are preserved.
	if got := genericCostAmount(result); got != 5 {
		t.Fatalf("stacked generic cost = %d, want 5", got)
	}
	if got := regColoredSymbolCount(result); got != 2 {
		t.Fatalf("colored symbols = %d, want 2 preserved", got)
	}
}

// TestRegCostReductionNeverGoesBelowZeroGeneric checks that an over-large
// reduction floors the generic portion at zero without touching colored mana.
func TestRegCostReductionNeverGoesBelowZeroGeneric(t *testing.T) {
	t.Parallel()
	base := cost.Mana{cost.O(2), cost.R}
	modifiers := []game.CostModifier{{GenericReduction: 5}}

	result := applyAbilityCostModifiers(&base, modifiers)

	if got := genericCostAmount(result); got != 0 {
		t.Fatalf("floored generic cost = %d, want 0", got)
	}
	if got := regColoredSymbolCount(result); got != 1 {
		t.Fatalf("colored symbols = %d, want 1 (the {R}) preserved", got)
	}
}

// TestRegCostMinimumGenericFloorApplies checks that a minimum-generic modifier
// raises a reduced cost back up to the floor.
func TestRegCostMinimumGenericFloorApplies(t *testing.T) {
	t.Parallel()
	base := cost.Mana{cost.O(4)}
	modifiers := []game.CostModifier{
		{GenericReduction: 3},
		{MinimumGeneric: 2},
	}

	result := applyAbilityCostModifiers(&base, modifiers)

	// 4 - 3 = 1, raised to the minimum of 2.
	if got := genericCostAmount(result); got != 2 {
		t.Fatalf("minimum-floored generic cost = %d, want 2", got)
	}
}

// TestRegCostSetGenericOverridesIncrements checks that a set-generic modifier
// replaces the accumulated increases and reductions.
func TestRegCostSetGenericOverridesIncrements(t *testing.T) {
	t.Parallel()
	base := cost.Mana{cost.O(5), cost.U}
	modifiers := []game.CostModifier{
		{GenericIncrease: 3},
		{SetGeneric: opt.Val(1)},
		{GenericReduction: 4},
	}

	result := applyAbilityCostModifiers(&base, modifiers)

	// The set value (1) overrides the +3/-4 increments entirely.
	if got := genericCostAmount(result); got != 1 {
		t.Fatalf("set-generic cost = %d, want 1", got)
	}
	if got := regColoredSymbolCount(result); got != 1 {
		t.Fatalf("colored symbols = %d, want 1 (the {U}) preserved", got)
	}
}

// TestRegCostModifierStackingProperty fuzzes many random increase/reduction/
// minimum combinations against an independent reference formula and the
// "never below zero generic" invariant.
func TestRegCostModifierStackingProperty(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewPCG(610, 9))
	for iteration := range 500 {
		baseGeneric := rng.IntN(7)
		base := cost.Mana{cost.G}
		if baseGeneric > 0 {
			base = cost.Mana{cost.O(baseGeneric), cost.G}
		}

		modifierCount := rng.IntN(5)
		modifiers := make([]game.CostModifier, 0, modifierCount)
		sumIncrease, sumReduction, minimum := 0, 0, 0
		for range modifierCount {
			switch rng.IntN(3) {
			case 0:
				inc := rng.IntN(4)
				sumIncrease += inc
				modifiers = append(modifiers, game.CostModifier{GenericIncrease: inc})
			case 1:
				red := rng.IntN(4)
				sumReduction += red
				modifiers = append(modifiers, game.CostModifier{GenericReduction: red})
			default:
				floor := rng.IntN(4)
				minimum = max(minimum, floor)
				modifiers = append(modifiers, game.CostModifier{MinimumGeneric: floor})
			}
		}

		want := baseGeneric + sumIncrease - sumReduction
		want = max(want, minimum)
		want = max(want, 0)

		result := applyAbilityCostModifiers(&base, modifiers)
		got := genericCostAmount(result)
		if got != want {
			t.Fatalf("iteration %d: generic = %d, want %d (base %d +%d -%d min %d)",
				iteration, got, want, baseGeneric, sumIncrease, sumReduction, minimum)
		}
		if got < 0 {
			t.Fatalf("iteration %d: generic cost dropped below zero: %d", iteration, got)
		}
		if regColoredSymbolCount(result) != 1 {
			t.Fatalf("iteration %d: colored {G} requirement was altered", iteration)
		}
	}
}
