package compiler

import "testing"

func TestCompileLinkedSearchRiderReference(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. Then if you control four or more lands, untap that land.",
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	content := compilation.Abilities[0].Content
	if len(content.Conditions) != 1 || !content.Conditions[0].Resolving {
		t.Fatalf("conditions = %#v, want resolving condition", content.Conditions)
	}
	if len(content.References) == 0 {
		t.Fatal("missing compiled references")
	}
	ref := content.References[len(content.References)-1]
	if ref.Kind != ReferenceThatObject ||
		ref.Binding != ReferenceBindingPriorInstructionResult ||
		ref.PriorInstruction != 0 {
		t.Fatalf("rider reference = %#v, want prior search result", ref)
	}
}
