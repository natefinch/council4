package compiler

import "testing"

// TestCompileOptionalIfYouDoFlow verifies the compiler types the "you may X.
// If you do, Y" flow as an optional first effect plus an affirmative
// prior-instruction condition contained in the second effect.
func TestCompileOptionalIfYouDoFlow(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"You may discard a card. If you do, draw two cards.",
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	content := compilation.Abilities[0].Content
	if len(content.Effects) != 2 {
		t.Fatalf("effects = %#v, want two", content.Effects)
	}
	if !content.Effects[0].Optional {
		t.Fatal("effect[0] optional = false, want optional discard")
	}
	if content.Effects[1].Optional {
		t.Fatal("effect[1] optional = true, want unconditional draw")
	}
	if len(content.Conditions) != 1 ||
		content.Conditions[0].Kind != ConditionIf ||
		content.Conditions[0].Predicate != ConditionPredicatePriorInstructionAccepted {
		t.Fatalf("conditions = %#v, want one if-you-do clause", content.Conditions)
	}
	if !content.Effects[1].Order.Contains(content.Conditions[0].Order) {
		t.Fatal("if-you-do condition not contained in the gated effect")
	}
}
