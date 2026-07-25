package cardgen

import (
	"testing"
)

// TestCompileCardDefsAppliesCorpusPolicyRefusals guards the equivalence the test
// seam depends on: CompileCardDefs must refuse exactly what production refuses.
//
// The refusals originally lived in the renderer-facing entry point, so the
// compile-only seam skipped them. A disowned card lowers perfectly well - being
// disowned is an identity decision, not a capability gap - so assertCardPaths
// would have produced a full dump and passed, claiming support for a card that
// is never generated. Worse, assertCardUnsupported could not test these refusals
// at all: it would see defs and no diagnostics and report the card as supported.
func TestCompileCardDefsAppliesCorpusPolicyRefusals(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		card *ScryfallCard
		want string
	}{
		{
			name: "disowned card",
			// Crusade lowers cleanly as an ordinary anthem, which is exactly
			// why the name-list refusal has to be reached before lowering.
			card: &ScryfallCard{
				Name:       "Crusade",
				Layout:     "normal",
				ManaCost:   "{W}{W}",
				TypeLine:   "Enchantment",
				OracleText: "White creatures get +1/+1.",
			},
			want: "disowned card excluded from generation",
		},
		{
			name: "unsupported layout",
			// cardLayoutValue maps an unknown layout to LayoutNormal, so
			// without this refusal the card would silently compile as though
			// it were a normal card.
			card: &ScryfallCard{
				Name:       "Unsupported Layout",
				Layout:     "art_series",
				TypeLine:   "Card",
				OracleText: "Draw a card.",
			},
			want: "unsupported card layout",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assertCardUnsupported(t, test.card, test.want)

			// Both entry points must agree, or the seam would again be
			// testing something production does not do.
			source, diagnostics, err := GenerateExecutableCardSource(test.card, "c")
			if err != nil {
				t.Fatalf("generate: %v", err)
			}
			if source != "" {
				t.Fatalf("generated source for a refused card:\n%s", source)
			}
			if len(diagnostics) != 1 || diagnostics[0].Summary != test.want {
				t.Fatalf("diagnostics = %#v, want one %q", diagnostics, test.want)
			}
		})
	}
}

// TestCompileCardDefsAcceptsSupportedCards is the other half: the refusals must
// not swallow an ordinary card. Without this, deleting the whole refusal check
// and deleting the whole compile path would both look the same to the test above.
func TestCompileCardDefsAcceptsSupportedCards(t *testing.T) {
	t.Parallel()
	defs, diagnostics, err := CompileCardDefs(&ScryfallCard{
		Name:       "Glorious Anthem",
		Layout:     "normal",
		ManaCost:   "{1}{W}{W}",
		TypeLine:   "Enchantment",
		OracleText: "Creatures you control get +1/+1.",
	})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diagnostics)
	}
	if len(defs) != 1 {
		t.Fatalf("defs = %d, want 1", len(defs))
	}
}
