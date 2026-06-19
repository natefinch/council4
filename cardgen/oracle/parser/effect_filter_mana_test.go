package parser

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/mana"
)

// TestParseFilterLandManaBodyExactness verifies that the filter-land output body
// "Add {X}{X}, {X}{Y}, or {Y}{Y}." parses into the typed FilterPair value and
// that near-miss bodies fail closed.
func TestParseFilterLandManaBodyExactness(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"{T}: Add {C}.\n{W/U}, {T}: Add {W}{W}, {W}{U}, or {U}{U}.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 2 {
		t.Fatalf("abilities = %d, want 2", len(document.Abilities))
	}
	filter := document.Abilities[1].Sentences[0].Effects[0].Mana
	if !filter.FilterPair {
		t.Fatalf("FilterPair = false, mana = %#v", filter)
	}
	if !slices.Equal(filter.FilterColors, []mana.Color{mana.W, mana.U}) {
		t.Fatalf("FilterColors = %#v, want [W U]", filter.FilterColors)
	}
	if !filter.LegacyBodyExact {
		t.Fatal("LegacyBodyExact = false, want true for an exact filter body")
	}
	if filter.Choice || filter.AnyColor || len(filter.Symbols) != 0 {
		t.Fatalf("filter body leaked into generic mana fields: %#v", filter)
	}

	for _, body := range []string{
		"{W/U}, {T}: Add {W}{W}, {W}{U}, or {U}.",       // final group only one mana
		"{W/U}, {T}: Add {W}{W} or {U}{U}.",             // only two groups
		"{W/U}, {T}: Add {W}{W}, {U}{W}, or {U}{U}.",    // mixed group is {U}{W} not {X}{Y}
		"{W/W}, {T}: Add {W}{W}, {W}{W}, or {W}{W}.",    // single color, not a pair
		"{C/U}, {T}: Add {C}{C}, {C}{U}, or {U}{U}.",    // colorless symbol in the pair
		"{W/U}, {T}: Add {W}{W}{W}, {W}{U}, or {U}{U}.", // first group three mana
		"{W/U}, {T}: Add {W}{W}, {W}{B}, or {U}{U}.",    // group2/group3 colors disagree
	} {
		near, _ := Parse("{T}: Add {C}.\n"+body, Context{})
		if len(near.Abilities) < 2 {
			continue
		}
		got := near.Abilities[1].Sentences[0].Effects[0].Mana
		if got.FilterPair {
			t.Fatalf("near-miss %q parsed as filter pair: %#v", body, got)
		}
	}
}
