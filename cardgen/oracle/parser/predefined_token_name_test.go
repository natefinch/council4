package parser

import "testing"

// TestParsePredefinedTokenName proves the "Create a [tapped] <Name> token." form
// for a predefined named token whose name is a card name rather than a card
// subtype (Mutavault) captures the name and parses exactly.
func TestParsePredefinedTokenName(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		source string
		name   string
		tapped bool
	}{
		{"Create a Mutavault token.", "Mutavault", false},
		{"Create a tapped Mutavault token.", "Mutavault", true},
	} {
		document, diagnostics := Parse(tc.source, Context{})
		if len(diagnostics) != 0 {
			t.Fatalf("Parse(%q) diagnostics = %#v", tc.source, diagnostics)
		}
		effect := &document.Abilities[0].Sentences[0].Effects[0]
		if effect.Kind != EffectCreate || !effect.Exact {
			t.Fatalf("Parse(%q) kind=%v exact=%v, want EffectCreate exact", tc.source, effect.Kind, effect.Exact)
		}
		if effect.TokenPredefinedName != tc.name {
			t.Fatalf("Parse(%q) TokenPredefinedName=%q, want %q", tc.source, effect.TokenPredefinedName, tc.name)
		}
		if effect.Selection.Tapped != tc.tapped {
			t.Fatalf("Parse(%q) Tapped=%v, want %v", tc.source, effect.Selection.Tapped, tc.tapped)
		}
	}
}

// TestParsePredefinedTokenNameFailsClosed proves that an unrecognized capitalized
// token noun is not captured as a predefined token name, so the create clause is
// not silently treated as a known token.
func TestParsePredefinedTokenNameFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Create a Spelltable token.",
		"Create a 1/1 white Soldier creature token.",
	} {
		document, diagnostics := Parse(source, Context{})
		if len(diagnostics) != 0 {
			t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
		}
		effect := &document.Abilities[0].Sentences[0].Effects[0]
		if effect.TokenPredefinedName != "" {
			t.Errorf("Parse(%q) captured predefined token name %q", source, effect.TokenPredefinedName)
		}
	}
}
