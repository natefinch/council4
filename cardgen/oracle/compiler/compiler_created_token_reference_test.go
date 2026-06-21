package compiler

import (
	"testing"
)

// TestCompileThatTokenBindsToCreatedToken verifies that the "that token"
// back-reference in "create a ... token. That token gains <keyword> ..." binds
// to the preceding token-creation effect's result rather than to a trigger-event
// permanent.
func TestCompileThatTokenBindsToCreatedToken(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"When this creature enters, create a 1/1 red Goblin creature token. That token gains haste until end of turn.",
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	content := compilation.Abilities[0].Content
	if len(content.Effects) != 2 {
		t.Fatalf("effects = %#v, want create then keyword grant", content.Effects)
	}
	if content.Effects[0].Kind != EffectCreate {
		t.Fatalf("effect[0] = %#v, want token creation", content.Effects[0])
	}
	grant := content.Effects[1]
	if len(grant.SubjectReferences) != 1 {
		t.Fatalf("grant subject references = %#v, want one", grant.SubjectReferences)
	}
	ref := grant.SubjectReferences[0]
	if ref.Kind != ReferenceThatObject ||
		ref.Binding != ReferenceBindingPriorInstructionResult ||
		ref.PriorInstruction != 0 {
		t.Fatalf("that-token reference = %#v, want prior create result", ref)
	}
}
