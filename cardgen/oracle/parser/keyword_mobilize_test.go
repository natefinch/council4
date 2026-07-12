package parser

import "testing"

// TestParseMobilizeFixedInteger proves the fixed "Mobilize N" form parses to the
// Mobilize keyword with an integer parameter carrying N.
func TestParseMobilizeFixedInteger(t *testing.T) {
	t.Parallel()
	keywords := keywordsFor(t, "Mobilize 2")
	if len(keywords) != 1 {
		t.Fatalf("keywords = %+v; want one", keywords)
	}
	if keywords[0].Kind != KeywordMobilize ||
		keywords[0].Parameter.Kind != KeywordParameterInteger ||
		keywords[0].Parameter.Integer() != 2 {
		t.Fatalf("mobilize = %+v", keywords[0])
	}
}

// TestParseMobilizeDynamicGraveyard proves the "Mobilize X, where X is the number
// of creature cards in your graveyard" form parses to the Mobilize keyword with
// the typed graveyard-count dynamic parameter, and that the keyword span covers
// the whole clause.
func TestParseMobilizeDynamicGraveyard(t *testing.T) {
	t.Parallel()
	source := "Mobilize X, where X is the number of creature cards in your graveyard"
	keywords := keywordsFor(t, source)
	if len(keywords) != 1 {
		t.Fatalf("keywords = %+v; want one", keywords)
	}
	if keywords[0].Kind != KeywordMobilize ||
		keywords[0].Parameter.Kind != KeywordParameterMobilizeDynamic ||
		keywords[0].Parameter.MobilizeDynamic() != MobilizeDynamicCreatureCardsInGraveyard {
		t.Fatalf("mobilize = %+v", keywords[0])
	}
	if keywords[0].Text != source {
		t.Fatalf("mobilize keyword text = %q; want the whole dynamic clause %q", keywords[0].Text, source)
	}
}

// TestParseMobilizeUnsupportedDynamicFailsClosed proves an unrecognized dynamic
// Mobilize form ("where X is its power") produces no integer or dynamic
// parameter and leaves the clause uncovered, so it fails closed downstream.
func TestParseMobilizeUnsupportedDynamicFailsClosed(t *testing.T) {
	t.Parallel()
	source := "Mobilize X, where X is its power"
	keywords := keywordsFor(t, source)
	if len(keywords) != 1 {
		t.Fatalf("keywords = %+v; want one", keywords)
	}
	if keywords[0].Kind != KeywordMobilize {
		t.Fatalf("mobilize kind = %v", keywords[0].Kind)
	}
	if keywords[0].Parameter.Kind == KeywordParameterMobilizeDynamic ||
		keywords[0].Parameter.Kind == KeywordParameterInteger {
		t.Fatalf("unsupported dynamic Mobilize produced a typed amount: %+v", keywords[0].Parameter)
	}
	// Only the bare keyword word is covered, so the "X, where X is its power"
	// clause remains uncovered for the coverage check and fails closed.
	if keywords[0].Text == source {
		t.Fatalf("mobilize keyword text = %q; want only the bare keyword word", keywords[0].Text)
	}
}
