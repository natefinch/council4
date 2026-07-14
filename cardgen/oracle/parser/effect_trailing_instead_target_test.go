package parser

import "testing"

func TestParseTrailingInsteadOutsideTarget(t *testing.T) {
	t.Parallel()
	source := "Exile target nonland permanent instead."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	effect := document.Abilities[0].Sentences[0].Effects[0]
	if len(effect.Targets) != 1 {
		t.Fatalf("targets = %#v, want one", effect.Targets)
	}
	target := effect.Targets[0]
	if !target.Exact || target.Text != "target nonland permanent" {
		t.Fatalf("target = %#v, want exact target without trailing instead", target)
	}
	if effect.Replacement.Kind != EffectReplacementInstead {
		t.Fatalf("replacement = %v, want EffectReplacementInstead", effect.Replacement.Kind)
	}
}
