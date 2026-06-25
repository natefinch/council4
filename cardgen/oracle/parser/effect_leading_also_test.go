package parser

import "testing"

// alsoLeadingCreateEffect parses a compound sentence whose second effect leads
// with the adverb "also" and returns that effect, so tests can assert the
// leading "also" does not disturb subject/exact classification.
func alsoLeadingCreateEffect(t *testing.T, source string) EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	var effects []EffectSyntax
	for _, sentence := range document.Abilities[0].Sentences {
		effects = append(effects, sentence.Effects...)
	}
	if len(effects) < 2 {
		t.Fatalf("Parse(%q) effects = %#v, want at least 2", source, effects)
	}
	return effects[len(effects)-1]
}

// TestExactLeadingAlsoCreateClause proves that a clause-leading "also" adverb is
// stripped during subject classification, so an additive "also create ..."
// sub-effect reconstructs byte-exactly and binds to the controller — matching
// the same clause without "also". This unlocks gated additive sub-effects such
// as Finale of Glory's "If X is 10 or more, also create ...".
func TestExactLeadingAlsoCreateClause(t *testing.T) {
	t.Parallel()
	withAlso := alsoLeadingCreateEffect(t, "Draw a card. Also create a 1/1 white Soldier creature token.")
	without := alsoLeadingCreateEffect(t, "Draw a card. Create a 1/1 white Soldier creature token.")

	if withAlso.Context != EffectContextController {
		t.Fatalf("leading-also effect context = %v, want EffectContextController", withAlso.Context)
	}
	if withAlso.Context != without.Context {
		t.Fatalf("leading-also context = %v, plain context = %v; want equal", withAlso.Context, without.Context)
	}
	if !withAlso.Exact {
		t.Fatal("leading-also create effect is not exact")
	}
	if withAlso.Exact != without.Exact {
		t.Fatalf("leading-also exact = %v, plain exact = %v; want equal", withAlso.Exact, without.Exact)
	}
}
