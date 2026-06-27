package parser

import (
	"slices"
	"testing"
)

// firstSearchEffect returns the first EffectSearch across all of a parse's
// sentences, so a two-sentence search ("Search ... . Put those cards ...") is
// reachable as well as a single-sentence one.
func firstSearchEffect(t *testing.T, source string) *EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("Parse(%q) abilities = %#v", source, document.Abilities)
	}
	for si := range document.Abilities[0].Sentences {
		effects := document.Abilities[0].Sentences[si].Effects
		for ei := range effects {
			if effects[ei].Kind == EffectSearch {
				return &effects[ei]
			}
		}
	}
	t.Fatalf("Parse(%q) has no EffectSearch", source)
	return nil
}

// TestDynamicCountLibrarySearchAccepts confirms the "up to X <filter>, where X is
// <rules-derived count>" search wording round-trips to a supported search effect
// in both the single-sentence ("..., put them ...") and two-sentence ("... . Put
// those cards ...") destination forms. The trailing "where X is ..." count phrase
// must be stripped from the reconstruction and must not bleed its counted-subject
// noun into the searched-card filter.
func TestDynamicCountLibrarySearchAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		// Single-sentence: the count phrase sits between the filter and the put.
		"Search your library for up to X basic land cards, where X is the number of lands you control, put them onto the battlefield tapped, then shuffle.",
		"Search your library for up to X basic land cards, where X is the number of tapped creatures you control, put those cards onto the battlefield tapped, then shuffle.",
		// Two-sentence: the search sentence ends after the count phrase; the put
		// is a separate following effect.
		"Search your library for up to X basic land cards, where X is the greatest power among creatures you control. Put those cards onto the battlefield tapped, then shuffle.",
		"Search your library for up to X basic land cards, where X is the number of colors among permanents you control. Reveal those cards, put them into your hand, then shuffle.",
	}
	for _, source := range accepted {
		search := firstSearchEffect(t, source)
		if search.UnsupportedDetail != "" {
			t.Errorf("UnsupportedDetail(%q) = %q, want empty", source, search.UnsupportedDetail)
		}
		// The searched-card filter is a basic-land selection; the count subject's
		// "creatures"/"permanents" noun must not bleed a creature type in.
		if slices.Contains(search.Selection.RequiredTypesAny, CardTypeCreature) {
			t.Errorf("Selection.RequiredTypesAny(%q) = %#v, contains creature (count subject bled in)", source, search.Selection.RequiredTypesAny)
		}
	}
}

// TestDynamicCountLibrarySearchFailsClosed confirms the dynamic-count search
// stays fail-closed when the "where X is ..." count is missing or names a
// rules-derived amount the runtime does not model, so an unrecognized "X" never
// lowers to a wrong count.
func TestDynamicCountLibrarySearchFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		// "up to X" with no "where X is ..." clause: X is the spell's {X}, which
		// this family does not yet model.
		"Search your library for up to X basic land cards, put them onto the battlefield tapped, then shuffle.",
		// An unrecognized rules-derived count ("creatures attacking you").
		"Search your library for up to X basic land cards, where X is the number of creatures attacking you, put those cards onto the battlefield tapped, then shuffle.",
	}
	for _, source := range rejected {
		search := firstSearchEffect(t, source)
		if search.UnsupportedDetail == "" {
			t.Errorf("UnsupportedDetail(%q) = empty, want non-empty (fail closed)", source)
		}
	}
}

// TestDynamicCountSearchFilterClean asserts the single-sentence bleed fix at the
// selection level: the where-X count subject ("tapped creatures you control")
// must not fold into the searched basic-land filter.
func TestDynamicCountSearchFilterClean(t *testing.T) {
	t.Parallel()
	source := "Search your library for up to X basic land cards, where X is the number of tapped creatures you control, put those cards onto the battlefield tapped, then shuffle."
	search := firstSearchEffect(t, source)
	if search.Selection.Kind != SelectionLand {
		t.Fatalf("Selection.Kind = %v, want SelectionLand", search.Selection.Kind)
	}
	if !slices.Equal(search.Selection.Supertypes, []Supertype{SupertypeBasic}) {
		t.Errorf("Selection.Supertypes = %#v, want [SupertypeBasic]", search.Selection.Supertypes)
	}
	if slices.Contains(search.Selection.RequiredTypesAny, CardTypeCreature) {
		t.Errorf("Selection.RequiredTypesAny = %#v, contains creature (count subject bled in)", search.Selection.RequiredTypesAny)
	}
}
