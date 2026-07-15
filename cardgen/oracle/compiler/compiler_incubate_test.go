package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// TestCompileIncubateStandaloneEffect proves the standalone imperative form
// lowers to an EffectIncubate effect scoped to the controller.
func TestCompileIncubateStandaloneEffect(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource("Incubate 2.", pipelineContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 1 ||
		effects[0].Kind != EffectIncubate ||
		effects[0].Context != parser.EffectContextController {
		t.Fatalf("effects = %#v, want controller incubate", effects)
	}
}

// TestCompileIncubateReferencedControllerEffect proves the Excise the Imperfect
// wording lowers its second sentence to an EffectIncubate scoped to the exiled
// permanent's controller, with a dynamic mana-value amount.
func TestCompileIncubateReferencedControllerEffect(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Exile target nonland permanent. Its controller incubates X, where X is its mana value.",
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 2 {
		t.Fatalf("effects = %#v, want two effects", effects)
	}
	if effects[1].Kind != EffectIncubate ||
		effects[1].Context != parser.EffectContextReferencedObjectController {
		t.Fatalf("effects[1] = %#v, want referenced-controller incubate", effects[1])
	}
}
