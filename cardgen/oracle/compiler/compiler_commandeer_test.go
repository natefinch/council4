package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompileCommandeerMechanics(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"You may exile two blue cards from your hand rather than pay this spell's mana cost.\n"+
			"Gain control of target noncreature spell. You may choose new targets for it. "+
			"(If that spell is an artifact, enchantment, or planeswalker, the permanent enters under your control.)",
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(compilation.Abilities) != 2 {
		t.Fatalf("abilities = %d, want 2", len(compilation.Abilities))
	}
	alternative := compilation.Abilities[0]
	if alternative.AlternativeCost == nil || alternative.AlternativeCost.Kind != AlternativeCostPitch {
		t.Fatalf("alternative cost = %#v, want pitch", alternative.AlternativeCost)
	}
	if alternative.Cost == nil || len(alternative.Cost.Components) != 1 ||
		alternative.Cost.Components[0].Kind != CostExile ||
		alternative.Cost.Components[0].AmountValue != 2 ||
		!alternative.Cost.Components[0].AmountKnown {
		t.Fatalf("compiled pitch cost = %#v, want exile two", alternative.Cost)
	}

	body := compilation.Abilities[1].Content
	if len(body.Targets) != 1 || len(body.Effects) != 2 {
		t.Fatalf("body = %#v, want one target and two effects", body)
	}
	target := body.Targets[0]
	if target.Selector.Kind != SelectorSpell ||
		len(target.Selector.ExcludedTypes()) != 1 ||
		target.Selector.ExcludedTypes()[0] != types.Creature {
		t.Fatalf("target = %#v, want noncreature spell", target)
	}
	retarget := body.Effects[1]
	if retarget.Kind != EffectChooseNewTargets || !retarget.Optional ||
		retarget.Context != parser.EffectContextController ||
		len(retarget.References) != 1 ||
		retarget.References[0].Binding != ReferenceBindingTarget ||
		retarget.References[0].Occurrence != 0 {
		t.Fatalf("retarget = %#v, want optional reference to target zero", retarget)
	}
}
