package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

// heterogeneousSlotSearchSyntax parses a "search for a X card and a Y card"
// library search and returns its leading search effect's per-slot subtypes
// alongside the generic-path fail-closed detail. The heterogeneous noun phrase is
// outside the byte-exact single-filter envelope, so the generic recognizer keeps
// its UnsupportedDetail set; the dedicated multi-slot lowering reads SearchSlots.
func heterogeneousSlotSearchSyntax(t *testing.T, source string) (slots []types.Sub, unsupported string) {
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
	return effects[0].SearchSlots, effects[0].UnsupportedDetail
}

// TestExactHeterogeneousSlotSearchAccepts confirms the Krosan Verge wording — two
// distinct single-subtype card slots joined by a plain "and" — records one
// basic-land subtype per slot in source order so the dedicated multi-slot
// lowering can collapse it, while the generic single-filter recognizer still
// fails closed (UnsupportedDetail set) so no other path mishandles it.
func TestExactHeterogeneousSlotSearchAccepts(t *testing.T) {
	t.Parallel()
	slots, unsupported := heterogeneousSlotSearchSyntax(t,
		"Search your library for a Forest card and a Plains card, put them onto the battlefield tapped, then shuffle.")
	if unsupported == "" {
		t.Fatal("heterogeneous slot search generic UnsupportedDetail = \"\", want the single-filter fail-closed detail")
	}
	want := []types.Sub{types.Forest, types.Plains}
	if len(slots) != len(want) {
		t.Fatalf("SearchSlots = %#v, want %#v", slots, want)
	}
	for i := range want {
		if slots[i] != want[i] {
			t.Fatalf("SearchSlots[%d] = %q, want %q", i, slots[i], want[i])
		}
	}
}

// TestHeterogeneousSlotSearchFailsClosed confirms wordings just outside the
// modeled two-slot heterogeneous shape carry no SearchSlots marker, so an
// ordinary single-filter search is never reinterpreted as a multi-slot one.
func TestHeterogeneousSlotSearchFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		// An "or" union is one slot of either subtype, not two distinct slots.
		"Search your library for a Forest or Plains card, put it onto the battlefield tapped, then shuffle.",
		// The two slots naming the same subtype is the homogeneous two-card case.
		"Search your library for a Forest card and a Forest card, put them onto the battlefield tapped, then shuffle.",
		// A non-basic-land subtype is outside the modeled basic-land slots.
		"Search your library for a Goblin card and a Forest card, put them onto the battlefield tapped, then shuffle.",
		// A plain card type carries no subtype slot to split on.
		"Search your library for a creature card and a land card, put them onto the battlefield tapped, then shuffle.",
		// A single basic-land search has only one slot.
		"Search your library for a Forest card, put it onto the battlefield tapped, then shuffle.",
		// Three slots are outside the modeled two-slot shape.
		"Search your library for a Forest card and a Plains card and an Island card, put them onto the battlefield tapped, then shuffle.",
	}
	for _, source := range rejected {
		if slots, _ := heterogeneousSlotSearchSyntax(t, source); len(slots) != 0 {
			t.Errorf("heterogeneous slot search %q SearchSlots = %#v, want none", source, slots)
		}
	}
}
