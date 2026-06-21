package parser

import "testing"

// TestParseAdditionalCardsDrawIsExact verifies that "draw N additional cards"
// (the extra-draw wording on draw-step triggers like Sylvan Library) parses as
// an exact fixed-count draw of N cards: the "additional" qualifier is captured
// by the Additional flag and restored in exact reconstruction, while the amount
// and selection are the same as a plain draw.
func TestParseAdditionalCardsDrawIsExact(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		value  int
	}{
		{"Draw two additional cards.", 2},
		{"Draw an additional card.", 1},
		{"You may draw three additional cards.", 3},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effect := document.Abilities[0].Sentences[0].Effects[0]
			if effect.Kind != EffectDraw ||
				!effect.Exact ||
				!effect.Additional ||
				!effect.Amount.Known ||
				effect.Amount.Value != test.value {
				t.Fatalf("effect = %#v, want exact additional draw of %d", effect, test.value)
			}
		})
	}
}

// TestParsePlainDrawNotAdditional guards that a plain draw is not marked
// Additional, so the reconstruction does not require the "additional" word.
func TestParsePlainDrawNotAdditional(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse("Draw two cards.", Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effect := document.Abilities[0].Sentences[0].Effects[0]
	if effect.Kind != EffectDraw || !effect.Exact || effect.Additional {
		t.Fatalf("effect = %#v, want exact non-additional draw", effect)
	}
}
