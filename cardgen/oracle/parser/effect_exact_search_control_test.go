package parser

import "testing"

// searchControlOf parses a single library-search sentence and returns its
// resolving search effect's exactness and recognized controller rider.
func searchControlOf(t *testing.T, source string) (bool, SearchControlRider) {
	t.Helper()
	document, diagnostics := Parse(source, Context{})
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
	return effects[0].Exact, effects[0].SearchControl
}

func TestExactLibrarySearchControlRiderAccepts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source  string
		control SearchControlRider
	}{
		{
			// Yavimaya Dryad: the found permanent enters under a chosen target
			// player's control rather than the searcher's.
			source:  "Search your library for a Forest card, put it onto the battlefield tapped under target player's control, then shuffle.",
			control: SearchControlRiderTargetPlayer,
		},
		{
			source:  "Search your library for a basic land card, put it onto the battlefield under target player's control, then shuffle.",
			control: SearchControlRiderTargetPlayer,
		},
		{
			source:  "Search your library for a creature card, put it onto the battlefield under target opponent's control, then shuffle.",
			control: SearchControlRiderTargetOpponent,
		},
	}
	for _, test := range tests {
		exact, control := searchControlOf(t, test.source)
		if !exact {
			t.Errorf("searchControlOf(%q) exact = false, want true", test.source)
		}
		if control != test.control {
			t.Errorf("searchControlOf(%q) control = %q, want %q", test.source, control, test.control)
		}
	}
}

func TestExactLibrarySearchControlRiderFailsClosed(t *testing.T) {
	t.Parallel()
	// The controller rider attaches only to a battlefield put. Every other
	// destination, and an unrecognized possessive controller, must fail closed
	// rather than silently drop the rider.
	rejected := []string{
		// A hand destination cannot carry the battlefield-only controller rider.
		"Search your library for a Forest card, put it into your hand under target player's control, then shuffle.",
		// A library-top destination cannot carry it either.
		"Search your library for a card, then shuffle and put that card on top under target player's control.",
		// "your control" / "its owner's control" are not target-player riders.
		"Search your library for a Forest card, put it onto the battlefield under your control, then shuffle.",
	}
	for _, source := range rejected {
		exact, control := searchControlOf(t, source)
		if exact {
			t.Errorf("searchControlOf(%q) exact = true, want false", source)
		}
		if control != SearchControlRiderNone {
			t.Errorf("searchControlOf(%q) control = %q, want none", source, control)
		}
	}
}
