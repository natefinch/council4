package parser

import "testing"

// TestParseLeadingNamedLegendaryToken proves the leading "Create <Name>, a
// legendary <P>/<T> <color> <Subtype> creature token with <keywords>." form
// (named legendary tokens such as Avacyn and Marit Lage) parses exactly, capturing
// the leading name, the legendary supertype, and every keyword.
func TestParseLeadingNamedLegendaryToken(t *testing.T) {
	t.Parallel()
	source := "Create Avacyn, a legendary 8/8 white Angel creature token with flying, vigilance, and indestructible."
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	effect := &document.Abilities[0].Sentences[0].Effects[0]
	if effect.Kind != EffectCreate || !effect.Exact {
		t.Fatalf("effect kind=%v exact=%v, want EffectCreate exact", effect.Kind, effect.Exact)
	}
	if effect.TokenName != "Avacyn" || !effect.TokenNameLeading {
		t.Fatalf("TokenName=%q leading=%v, want \"Avacyn\" leading", effect.TokenName, effect.TokenNameLeading)
	}
	if len(effect.Selection.Supertypes) != 1 || effect.Selection.Supertypes[0] != SupertypeLegendary {
		t.Fatalf("Supertypes=%v, want [Legendary]", effect.Selection.Supertypes)
	}
	if len(effect.TokenKeywords) != 3 {
		t.Fatalf("TokenKeywords=%v, want three keywords", effect.TokenKeywords)
	}
}

// TestParseLeadingTokenNameFailsClosed proves create wordings without the leading
// "<Name>, a" shape do not capture a spurious leading token name.
func TestParseLeadingTokenNameFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Create a 1/1 white Soldier creature token.",
		"Create two 1/1 white Soldier creature tokens.",
	} {
		document, diagnostics := Parse(source, Context{})
		if len(diagnostics) != 0 {
			t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
		}
		effect := &document.Abilities[0].Sentences[0].Effects[0]
		if effect.TokenNameLeading || effect.TokenName != "" {
			t.Errorf("Parse(%q) captured leading name %q (leading=%v)", source, effect.TokenName, effect.TokenNameLeading)
		}
	}
}

// TestParseCommaJoinedNegatedTypeTarget proves a target noun phrase whose filter
// joins two negated card types with a comma ("non-Saga, nonland permanent") keeps
// the whole phrase on the target rather than truncating at the internal comma.
func TestParseCommaJoinedNegatedTypeTarget(t *testing.T) {
	t.Parallel()
	source := "Exile target non-Saga, nonland permanent."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	effect := &document.Abilities[0].Sentences[0].Effects[0]
	if len(effect.Targets) != 1 {
		t.Fatalf("Targets = %#v, want one", effect.Targets)
	}
	if got := effect.Targets[0].Text; got != "target non-Saga, nonland permanent" {
		t.Fatalf("target text = %q, want full comma-joined phrase", got)
	}
	if len(effect.Selection.ExcludedTypes) != 1 || effect.Selection.ExcludedTypes[0] != CardTypeLand {
		t.Fatalf("ExcludedTypes = %v, want [Land]", effect.Selection.ExcludedTypes)
	}
}
