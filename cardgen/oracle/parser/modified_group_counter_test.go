package parser

import "testing"

func TestParseModifiedCreatureGroupCounterPlacement(t *testing.T) {
	t.Parallel()
	source := "Put a +1/+1 counter on each modified creature you control."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effect := document.Abilities[0].Sentences[0].Effects[0]
	if !effect.Selection.Modified || !effect.Exact {
		object, objectOK := exactGroupDamagePermanentRecipientText(effect.Selection)
		t.Fatalf("modified = %v, exact = %v, object = %q (ok=%v), selection = %#v",
			effect.Selection.Modified, effect.Exact, object, objectOK, effect.Selection)
	}
}
