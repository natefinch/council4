package parser

import (
	"testing"
)

// sharedSubtypeSearchSyntax parses a correlated "that share a land type" library
// search and returns its leading search effect's exact round-trip status
// alongside the SearchSharedSubtype correlation flag the parser recorded on it.
func sharedSubtypeSearchSyntax(t *testing.T, source string) (exact, sharedSubtype bool) {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) == 0 || effects[0].Kind != EffectSearch {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0].Exact, effects[0].SearchSharedSubtype
}

func TestExactSharedLandTypeSearchAccepts(t *testing.T) {
	t.Parallel()
	// Myriad Landscape and the safe wording neighbors that keep the same
	// two-card basic-land correlation: both found cards must share a land
	// subtype, and they reach a hand or battlefield (optionally tapped,
	// optionally revealed) destination together.
	accepted := []string{
		// Myriad Landscape's exact activation effect.
		"Search your library for up to two basic land cards that share a land type, put them onto the battlefield tapped, then shuffle.",
		"Search your library for up to two basic land cards that share a land type, put them onto the battlefield, then shuffle.",
		"Search your library for up to two basic land cards that share a land type, put them into your hand, then shuffle.",
		"Search your library for up to two basic land cards that share a land type, reveal those cards, put them into your hand, then shuffle.",
	}
	for _, source := range accepted {
		exact, sharedSubtype := sharedSubtypeSearchSyntax(t, source)
		if !exact {
			t.Errorf("shared-land-type search %q exact = false, want true", source)
		}
		if !sharedSubtype {
			t.Errorf("shared-land-type search %q SearchSharedSubtype = false, want true", source)
		}
	}
}

func TestExactSharedLandTypeSearchFailsClosed(t *testing.T) {
	t.Parallel()
	// Each wording sits just outside the modeled two-card basic-land correlation,
	// so the search must fail closed rather than drop the correlation silently.
	rejected := []string{
		// A singular search has no second card to correlate with.
		"Search your library for a basic land card that share a land type, put it onto the battlefield tapped, then shuffle.",
		// More than two cards is outside the modeled two-card shape.
		"Search your library for up to three basic land cards that share a land type, put them onto the battlefield tapped, then shuffle.",
		// "share a land type" is not meaningful for a non-land filter.
		"Search your library for up to two creature cards that share a land type, put them onto the battlefield tapped, then shuffle.",
		// A different correlation property is not modeled.
		"Search your library for up to two basic land cards that share a color, put them onto the battlefield tapped, then shuffle.",
		// An anti-correlation ("don't share") is not the modeled rider.
		"Search your library for up to two basic land cards that don't share a land type, put them onto the battlefield tapped, then shuffle.",
		// A graveyard destination is not a modeled correlated-search destination.
		"Search your library for up to two basic land cards that share a land type, put them into your graveyard, then shuffle.",
		// An extra trailing clause breaks the byte-exact envelope.
		"Search your library for up to two basic land cards that share a land type, put them onto the battlefield tapped, then shuffle, then draw a card.",
	}
	for _, source := range rejected {
		if exact, _ := sharedSubtypeSearchSyntax(t, source); exact {
			t.Errorf("shared-land-type search %q exact = true, want false", source)
		}
	}
}

// TestExactPlainSearchHasNoSharedSubtypeFlag confirms an ordinary basic-land
// search without the correlation rider parses exact but leaves the
// SearchSharedSubtype flag unset, so unrelated tutors keep their semantics.
func TestExactPlainSearchHasNoSharedSubtypeFlag(t *testing.T) {
	t.Parallel()
	exact, sharedSubtype := sharedSubtypeSearchSyntax(t,
		"Search your library for up to two basic land cards, put them onto the battlefield tapped, then shuffle.")
	if !exact {
		t.Fatal("plain two-card basic-land search exact = false, want true")
	}
	if sharedSubtype {
		t.Fatal("plain two-card basic-land search set SearchSharedSubtype, want unset")
	}
}
