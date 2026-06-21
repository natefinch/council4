package parser

import "testing"

// exileTopEffect parses a single exile sentence and returns its resolving
// effect for inspection of the exile-top-of-library recognizer.
func exileTopEffect(t *testing.T, source string) *EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectExile {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return &effects[0]
}

func TestExactExileTopOfLibraryAccepts(t *testing.T) {
	t.Parallel()
	cases := []struct {
		source string
		amount int
	}{
		{"Exile the top card of your library.", 1},
		{"Exile the top three cards of your library.", 3},
		{"Each player exiles the top three cards of their library.", 3},
		{"Each opponent exiles the top two cards of their library.", 2},
		{"Target opponent exiles the top card of their library.", 1},
	}
	for _, c := range cases {
		effect := exileTopEffect(t, c.source)
		if !effect.Exact {
			t.Errorf("exileTopEffect(%q).Exact = false, want true", c.source)
		}
		if effect.CardSource != EffectCardSourceTopOfPlayerLibrary {
			t.Errorf("exileTopEffect(%q).CardSource = %q, want %q",
				c.source, effect.CardSource, EffectCardSourceTopOfPlayerLibrary)
		}
		if !effect.Amount.Known || effect.Amount.Value != c.amount {
			t.Errorf("exileTopEffect(%q).Amount = %+v, want known %d",
				c.source, effect.Amount, c.amount)
		}
	}
}

func TestExactExileTopOfLibraryRejects(t *testing.T) {
	t.Parallel()
	// "Exile the top card of target player's library" is a different zone owner
	// the recognizer does not reconstruct, and a plain targeted permanent exile
	// must not be misread as a top-of-library source.
	cases := []string{
		"Exile target creature.",
		"Exile target permanent.",
	}
	for _, source := range cases {
		effect := exileTopEffect(t, source)
		if effect.CardSource == EffectCardSourceTopOfPlayerLibrary {
			t.Errorf("exileTopEffect(%q).CardSource = top-of-library, want none", source)
		}
	}
}
