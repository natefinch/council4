package parser

import (
	"testing"
)

// differentNamesSearchSyntax parses a correlated "with different names" library
// search and returns its leading search effect's exact round-trip status
// alongside the SearchDifferentNames correlation flag the parser recorded on it.
func differentNamesSearchSyntax(t *testing.T, source string) (exact, differentNames bool) {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) == 0 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) == 0 || effects[0].Kind != EffectSearch {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0].Exact, effects[0].SearchDifferentNames
}

func TestExactDifferentNamesSearchAccepts(t *testing.T) {
	t.Parallel()
	// Multi-card tutors whose found cards must each have a distinct name reach a
	// hand or battlefield destination together (Three Dreams, Shared Summons,
	// Deathbellow War Cry). The fixed and "up to X" counts both qualify.
	accepted := []string{
		"Search your library for up to three Aura cards with different names, reveal them, put them into your hand, then shuffle.",
		"Search your library for up to two creature cards with different names, put them onto the battlefield, then shuffle.",
		"Search your library for up to four Minotaur creature cards with different names, put them onto the battlefield, then shuffle.",
	}
	for _, source := range accepted {
		exact, differentNames := differentNamesSearchSyntax(t, source)
		if !exact {
			t.Errorf("different-names search %q exact = false, want true", source)
		}
		if !differentNames {
			t.Errorf("different-names search %q SearchDifferentNames = false, want true", source)
		}
	}
}

func TestExactDifferentNamesSearchFailsClosed(t *testing.T) {
	t.Parallel()
	// A singular search has no second card to correlate with, so the rider is not
	// meaningful and the wording must fail closed rather than lower silently.
	rejected := []string{
		"Search your library for a creature card with different names, put it onto the battlefield, then shuffle.",
	}
	for _, source := range rejected {
		if exact, _ := differentNamesSearchSyntax(t, source); exact {
			t.Errorf("different-names search %q exact = true, want false", source)
		}
	}
}

// TestExactPlainSearchHasNoDifferentNamesFlag confirms an ordinary multi-card
// search without the correlation rider parses exact but leaves the
// SearchDifferentNames flag unset, so unrelated tutors keep their semantics.
func TestExactPlainSearchHasNoDifferentNamesFlag(t *testing.T) {
	t.Parallel()
	exact, differentNames := differentNamesSearchSyntax(t,
		"Search your library for up to two creature cards, put them onto the battlefield, then shuffle.")
	if !exact {
		t.Fatal("plain two-card creature search exact = false, want true")
	}
	if differentNames {
		t.Fatal("plain two-card creature search set SearchDifferentNames, want unset")
	}
}
