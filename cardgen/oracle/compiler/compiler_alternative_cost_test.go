package compiler

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
)

func TestCompileCommanderControlledAlternativeSpellCost(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"If you control a commander, you may cast this spell without paying its mana cost.\nCounter target noncreature spell.",
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if ability.Kind != AbilitySpellAlternativeCost || ability.AlternativeCost == nil {
		t.Fatalf("ability = %#v, want typed alternative cost", ability)
	}

	if ability.AlternativeCost.Condition != AlternativeCostConditionControlsCommander ||
		!ability.AlternativeCost.WithoutPayingManaCost {
		t.Fatalf("alternative cost = %#v", ability.AlternativeCost)
	}
	if len(ability.Content.Effects) != 0 || len(ability.Content.Conditions) != 0 {
		t.Fatalf("alternative cost produced resolving content: %#v", ability.Content)
	}
}

func TestCompileOverloadAlternativeSpellCost(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		`Destroy target artifact you don't control.
Overload {4}{R} (You may cast this spell for its overload cost. If you do, change "target" in its text to "each.")`,
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[1]
	if ability.AlternativeCost == nil ||
		ability.AlternativeCost.Kind != AlternativeCostOverload ||
		!ability.AlternativeCost.ReplaceTargetWithEach ||
		!slices.Equal(ability.AlternativeCost.ManaCost, cost.Mana{cost.O(4), cost.R}) {
		t.Fatalf("alternative cost = %#v", ability.AlternativeCost)
	}
}
