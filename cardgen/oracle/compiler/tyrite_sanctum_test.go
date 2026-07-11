package compiler

import "testing"

func TestCompilePermanentSubtypeThenCounter(t *testing.T) {
	t.Parallel()

	source := "{2}, {T}: Target legendary creature becomes a God in addition to its other types. Put a +1/+1 counter on it."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	content := compilation.Abilities[0].Content
	if len(content.Effects) != 2 {
		t.Fatalf("effects = %#v, want become-type then counter", content.Effects)
	}
	first := content.Effects[0]
	if len(first.BecomeTypeAddSubtypes) != 1 ||
		len(first.Targets) != 1 || len(first.References) != 1 {
		t.Fatalf("become-type effect = %#v, want subtype, target, and possessive reference", first)
	}
	second := content.Effects[1]
	if len(second.References) != 1 {
		t.Fatalf("counter effect = %#v, want target back-reference", second)
	}
}
