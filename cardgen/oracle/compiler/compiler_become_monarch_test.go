package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// TestCompileBecomeMonarchControllerEffect proves the controller form lowers to
// an EffectBecomeMonarch effect scoped to the controller.
func TestCompileBecomeMonarchControllerEffect(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource("You become the monarch.", pipelineContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 1 ||
		effects[0].Kind != EffectBecomeMonarch ||
		effects[0].Context != parser.EffectContextController {
		t.Fatalf("effects = %#v, want controller become-monarch", effects)
	}
}

// TestCompileBecomeMonarchTargetEffect proves the single player-target form
// lowers to an EffectBecomeMonarch effect scoped to the chosen target.
func TestCompileBecomeMonarchTargetEffect(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource("Target player becomes the monarch.", pipelineContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 1 ||
		effects[0].Kind != EffectBecomeMonarch ||
		effects[0].Context != parser.EffectContextTarget {
		t.Fatalf("effects = %#v, want target become-monarch", effects)
	}
}
