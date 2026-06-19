package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

// splitPutSyntax parses a split-destination land tutor and returns its leading
// search effect's exact round-trip status alongside the structured split slots
// the parser recorded on the EffectPut clause.
func splitPutSyntax(t *testing.T, source string) (exact bool, split SearchSplitSyntax) {
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
	for i := range effects {
		if effects[i].Kind == EffectPut {
			split = effects[i].SearchSplit
		}
	}
	return effects[0].Exact, split
}

func TestExactSplitDestinationSearchAccepts(t *testing.T) {
	t.Parallel()
	// Cultivate and Kodama's Reach: search for up to two basic lands, reveal
	// them, then split one onto the battlefield tapped and the other to hand.
	tests := []struct {
		source string
		first  SearchSplitSlot
		second SearchSplitSlot
	}{
		{
			source: "Search your library for up to two basic land cards, reveal those cards, put one onto the battlefield tapped and the other into your hand, then shuffle.",
			first:  SearchSplitSlot{ToZone: zone.Battlefield, EntersTapped: true},
			second: SearchSplitSlot{ToZone: zone.Hand},
		},
		{
			source: "Search your library for up to two basic land cards, reveal them, put one onto the battlefield tapped and the other into your hand, then shuffle.",
			first:  SearchSplitSlot{ToZone: zone.Battlefield, EntersTapped: true},
			second: SearchSplitSlot{ToZone: zone.Hand},
		},
		{
			source: "Search your library for up to two basic land cards, put one onto the battlefield tapped and the other into your hand, then shuffle.",
			first:  SearchSplitSlot{ToZone: zone.Battlefield, EntersTapped: true},
			second: SearchSplitSlot{ToZone: zone.Hand},
		},
		{
			source: "Search your library for up to two basic land cards, reveal those cards, put one into your hand and the other onto the battlefield tapped, then shuffle.",
			first:  SearchSplitSlot{ToZone: zone.Hand},
			second: SearchSplitSlot{ToZone: zone.Battlefield, EntersTapped: true},
		},
	}
	for _, test := range tests {
		exact, split := splitPutSyntax(t, test.source)
		if !exact {
			t.Errorf("split search %q exact = false, want true", test.source)
		}
		if !split.Present {
			t.Errorf("split search %q SearchSplit.Present = false, want true", test.source)
			continue
		}
		if split.First != test.first || split.Second != test.second {
			t.Errorf("split search %q slots = %+v/%+v, want %+v/%+v", test.source, split.First, split.Second, test.first, test.second)
		}
	}
}

func TestExactSplitDestinationSearchFailsClosed(t *testing.T) {
	t.Parallel()
	// Every wording carries a rider outside the modeled two-slot split envelope,
	// so the search must fail closed rather than lower to a wrong distribution.
	rejected := []string{
		// More than two cards cannot fill exactly two single-card slots.
		"Search your library for up to three basic land cards, reveal those cards, put one onto the battlefield tapped and the other into your hand, then shuffle.",
		// A singular search has no second card for "the other".
		"Search your library for a basic land card, put one onto the battlefield tapped and the other into your hand, then shuffle.",
		// Myriad Landscape's "that share a land type" constraint is not modeled.
		"Search your library for up to two basic land cards that share a land type, reveal those cards, put one onto the battlefield tapped and the other into your hand, then shuffle.",
		// A third destination slot exceeds the two-slot model.
		"Search your library for up to two basic land cards, put one onto the battlefield tapped, one into your hand, and the other into your graveyard, then shuffle.",
		// A graveyard slot is not a modeled split destination.
		"Search your library for up to two basic land cards, put one onto the battlefield tapped and the other into your graveyard, then shuffle.",
		// An extra trailing clause after the split breaks the byte-exact envelope.
		"Search your library for up to two basic land cards, put one onto the battlefield tapped and the other into your hand, then shuffle, then draw a card.",
	}
	for _, source := range rejected {
		if exact, _ := splitPutSyntax(t, source); exact {
			t.Errorf("split search %q exact = true, want false", source)
		}
	}
}
