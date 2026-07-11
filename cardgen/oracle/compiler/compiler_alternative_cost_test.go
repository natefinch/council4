package compiler

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
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

func TestCompileCommanderControlledCreatureExile(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"If you control a commander, you may cast this spell without paying its mana cost.\nExile target creature.",
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(compilation.Abilities) != 2 {
		t.Fatalf("abilities = %d, want 2", len(compilation.Abilities))
	}
	alternative, spell := compilation.Abilities[0], compilation.Abilities[1]
	if alternative.AlternativeCost == nil ||
		alternative.AlternativeCost.Kind != AlternativeCostCommander ||
		alternative.AlternativeCost.Condition != AlternativeCostConditionControlsCommander ||
		!alternative.AlternativeCost.WithoutPayingManaCost {
		t.Fatalf("alternative cost = %#v", alternative.AlternativeCost)
	}
	if len(alternative.Content.Effects) != 0 ||
		len(alternative.Content.Targets) != 0 ||
		len(alternative.Content.References) != 0 {
		t.Fatalf("alternative content = %#v, want empty", alternative.Content)
	}
	if len(spell.Content.Effects) != 1 || len(spell.Content.Targets) != 1 {
		t.Fatalf("spell content = %#v", spell.Content)
	}
	effect, target := spell.Content.Effects[0], spell.Content.Targets[0]
	if effect.Kind != EffectExile || !effect.Exact || len(effect.References) != 0 {
		t.Fatalf("effect = %#v, want exact reference-free exile", effect)
	}
	if !target.Exact ||
		target.Cardinality != (TargetCardinality{Min: 1, Max: 1}) ||
		target.Selector.Kind != SelectorCreature {
		t.Fatalf("target = %#v, want exact one creature", target)
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

func TestCompileFreeAlternativeSpellCost(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"If you control a Swamp, you may pay 4 life rather than pay this spell's mana cost.\nDestroy target nonblack creature. It can't be regenerated.",
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.AlternativeCost == nil ||
		ability.AlternativeCost.Kind != AlternativeCostFree ||
		ability.AlternativeCost.Condition != AlternativeCostConditionControlsSubtype ||
		ability.AlternativeCost.ConditionSubtype != types.Swamp ||
		ability.AlternativeCost.WithoutPayingManaCost {
		t.Fatalf("alternative cost = %#v", ability.AlternativeCost)
	}
	if ability.Cost == nil || len(ability.Cost.Components) != 1 {
		t.Fatalf("cost = %#v, want one pay-life component", ability.Cost)
	}
	if ability.Cost.Components[0].Kind != CostPayLife {
		t.Fatalf("cost component = %#v, want pay life", ability.Cost.Components[0])
	}
}
