package parser

import "testing"

// TestParseIncubateClassifiesVerb proves the incubate keyword-action verb
// classifies to EffectIncubate in both the standalone imperative form and the
// referenced-object-controller form.
func TestParseIncubateClassifiesVerb(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source        string
		sentence      int
		effectContext EffectContextKind
	}{
		{"Incubate 2.", 0, EffectContextController},
		{
			"Exile target nonland permanent. Its controller incubates X, where X is its mana value.",
			1,
			EffectContextReferencedObjectController,
		},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[test.sentence].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want exactly one", effects)
			}
			if effects[0].Kind != EffectIncubate {
				t.Fatalf("kind = %v, want EffectIncubate", effects[0].Kind)
			}
			if effects[0].Context != test.effectContext {
				t.Fatalf("context = %v, want %v", effects[0].Context, test.effectContext)
			}
		})
	}
}

// TestParseIncubateExactness proves the incubate exactness recognizer accepts
// the exact standalone and referenced-controller wordings (including the dynamic
// "where X is its mana value" form) and fails closed for near-miss wordings.
func TestParseIncubateExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source   string
		sentence int
		exact    bool
	}{
		{"Incubate 2.", 0, true},
		{
			"Exile target nonland permanent. Its controller incubates X, where X is its mana value.",
			1,
			true,
		},
		{
			"Destroy target creature. Its controller incubates 3.",
			1,
			true,
		},
		// The "its power" dynamic wording is a well-formed incubate clause too;
		// parser exactness is text-faithful (the mana-value restriction is a
		// lowering decision, not a parser one).
		{
			"Exile target nonland permanent. Its controller incubates X, where X is its power.",
			1,
			true,
		},
		// A singular-verb near-miss ("incubate" not "incubates") does not
		// reconstruct byte-for-byte and fails closed.
		{
			"Exile target nonland permanent. Its controller incubate X, where X is its mana value.",
			1,
			false,
		},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[test.sentence].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want exactly one", effects)
			}
			if effects[0].Kind != EffectIncubate {
				t.Fatalf("kind = %v, want EffectIncubate", effects[0].Kind)
			}
			if effects[0].Exact != test.exact {
				t.Fatalf("Exact = %v, want %v", effects[0].Exact, test.exact)
			}
		})
	}
}
