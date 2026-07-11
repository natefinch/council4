package compiler

import "testing"

func TestCompileTappedTargetCostReduction(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"This spell costs {2} less to cast if it targets a tapped creature.",
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effect := compilation.Abilities[0].Content.Effects[0]
	if !effect.SourceSpellCostReductionConditional ||
		!effect.SourceSpellCostReductionTargetsTappedCreature ||
		effect.SourceSpellCostReductionAmount != 2 {
		t.Fatalf("effect = %#v, want typed tapped-target reduction", effect)
	}
}
