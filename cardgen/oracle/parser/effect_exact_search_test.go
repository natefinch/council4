package parser

import "testing"

// searchExact parses a single library-search sentence and reports whether its
// resolving effect round-tripped to an exact, lowerable production.
func searchExact(t *testing.T, source string) bool {
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
	return effects[0].Exact
}

func TestExactLibrarySearchAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		// Plain and single card-type singular searches.
		"Search your library for a card, put that card into your hand, then shuffle.",
		"Search your library for a creature card, reveal it, put it into your hand, then shuffle.",
		"Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
		"Search your library for a land card, put it onto the battlefield tapped, then shuffle.",
		// Basic land subtype unions (fetch and dual lands).
		"Search your library for a Forest or Island card, put it onto the battlefield, then shuffle.",
		"Search your library for a Mountain or Forest card, put it onto the battlefield, then shuffle.",
		"Search your library for a basic Forest card, put it onto the battlefield tapped, then shuffle.",
		"Search your library for a basic Forest, Plains, or Island card, put it onto the battlefield tapped, then shuffle.",
		// "up to N" plural searches with plural destination wording.
		"Search your library for up to two basic land cards, put them onto the battlefield tapped, then shuffle.",
		"Search your library for up to three basic land cards, put them onto the battlefield, then shuffle.",
		"Search your library for up to two enchantment cards, reveal them, put them into your hand, then shuffle.",
		"Search your library for up to two basic land cards, put those cards onto the battlefield tapped, then shuffle.",
		// Non-basic subtype searches (the subtype implies the card type).
		"Search your library for a Sliver card, reveal it, put it into your hand, then shuffle.",
		"Search your library for an Equipment card, put it onto the battlefield, then shuffle.",
		"Search your library for an Aura or Equipment card, put it into your hand, then shuffle.",
		// A subtype paired with a card type.
		"Search your library for a Myr creature card, put it onto the battlefield, then shuffle.",
		"Search your library for a Dragon creature card, reveal it, put it into your hand, then shuffle.",
		// Planeswalker tutors, singular and "up to N".
		"Search your library for a planeswalker card, reveal it, put it into your hand, then shuffle.",
		"Search your library for up to two planeswalker cards, reveal them, put them into your hand, then shuffle.",
	}
	for _, source := range accepted {
		if !searchExact(t, source) {
			t.Errorf("searchExact(%q) = false, want true", source)
		}
	}
}

func TestExactLibrarySearchFailsClosed(t *testing.T) {
	t.Parallel()
	// Each carries a rider the runtime SearchSpec cannot faithfully express, so
	// the round-trip must fail closed rather than lower to a wrong predicate.
	rejected := []string{
		// Non-library or extra source zone.
		"Search your library and graveyard for a creature card, put it into your hand, then shuffle.",
		// Color and mana-value filters are not modeled.
		"Search your library for a green creature card, put it onto the battlefield, then shuffle.",
		"Search your library for a creature card with mana value 3 or less, put it into your hand, then shuffle.",
		// Instant and sorcery reach the parser as a card kind carrying a required
		// card type the compiler drops, so the lowered spec would silently lose
		// the type; they must fail closed.
		"Search your library for an instant card, reveal it, put it into your hand, then shuffle.",
		"Search your library for an instant or sorcery card, reveal it, put it into your hand, then shuffle.",
		// "permanent" is not a modeled card type, and a multi-type union exceeds
		// the single-type SearchSpec.
		"Search your library for a Goblin permanent card, put it onto the battlefield, then shuffle.",
		"Search your library for an artifact creature card, put it onto the battlefield, then shuffle.",
		// "different names" and variable counts.
		"Search your library for up to two basic land cards with different names, put them onto the battlefield tapped, then shuffle.",
		"Search your library for up to X basic land cards, put them onto the battlefield tapped, then shuffle.",
		// Unsupported destinations.
		"Search your library for a card, put that card into your graveyard, then shuffle.",
		"Search your library for a card, put it on top of your library, then shuffle.",
	}
	for _, source := range rejected {
		if searchExact(t, source) {
			t.Errorf("searchExact(%q) = true, want false", source)
		}
	}
}

// searchExactOptional parses a single optional library-search sentence ("You may
// search ...") and reports both that its leading search effect carries the
// resolving optionality and whether it round-tripped to an exact production. The
// "you may" prefix must not defeat exact recognition.
func searchExactOptional(t *testing.T, source string) (optional, exact bool) {
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
	return effects[0].Optional, effects[0].Exact
}

func TestExactOptionalLibrarySearchAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"You may search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
		"You may search your library for a creature card, reveal it, put it into your hand, then shuffle.",
		"You may search your library for a Goblin card, reveal it, put it into your hand, then shuffle.",
		"You may search your library for up to two basic land cards, put them onto the battlefield tapped, then shuffle.",
	}
	for _, source := range accepted {
		optional, exact := searchExactOptional(t, source)
		if !optional {
			t.Errorf("searchExactOptional(%q) optional = false, want true", source)
		}
		if !exact {
			t.Errorf("searchExactOptional(%q) exact = false, want true", source)
		}
	}
}

func TestExactOptionalLibrarySearchFailsClosed(t *testing.T) {
	t.Parallel()
	// The optional prefix must not relax the filter/shape envelope: an
	// unsupported filter stays non-exact even when wrapped in "you may".
	rejected := []string{
		"You may search your library for an instant card, reveal it, put it into your hand, then shuffle.",
		"You may search your library and graveyard for a creature card, put it into your hand, then shuffle.",
	}
	for _, source := range rejected {
		if _, exact := searchExactOptional(t, source); exact {
			t.Errorf("searchExactOptional(%q) exact = true, want false", source)
		}
	}
}
