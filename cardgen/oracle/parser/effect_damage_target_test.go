package parser

import "testing"

// damageEffectExact parses a single self-name damage sentence and reports
// whether its resolving effect round-tripped to an exact, lowerable production.
func damageEffectExact(t *testing.T, name, source string) bool {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true, CardName: name})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectDealDamage {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0].Exact
}

func TestExactDamageTargetAccepts(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, source string }{
		{"Lava Spike", "Lava Spike deals 3 damage to target player or planeswalker."},
		{"Searing Flesh", "Searing Flesh deals 7 damage to target opponent or planeswalker."},
		{"Leaf Arrow", "Leaf Arrow deals 3 damage to target creature with flying."},
		{"Rending Volley", "Rending Volley deals 4 damage to target white or blue creature."},
		{"Gale Force", "Gale Force deals 5 damage to each creature with flying."},
	}
	for _, test := range tests {
		if !damageEffectExact(t, test.name, test.source) {
			t.Errorf("damageEffectExact(%q) = false, want true", test.source)
		}
	}
}

func TestExactDamageTargetFailsClosed(t *testing.T) {
	t.Parallel()
	// "without flying" uses an excluded keyword that SelectionSyntax does not
	// capture, so the recipient cannot round-trip and stays fail-closed.
	tests := []struct{ name, source string }{
		{"Antiflyer", "Antiflyer deals 1 damage to each creature without flying."},
	}
	for _, test := range tests {
		if damageEffectExact(t, test.name, test.source) {
			t.Errorf("damageEffectExact(%q) = true, want false", test.source)
		}
	}
}
