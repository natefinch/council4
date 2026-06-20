package compiler

import "testing"

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
