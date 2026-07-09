package parser

import (
	"testing"
)

// TestParseSearchManaValueDynamicCountBound proves the parser recognizes a
// library search whose searched-card mana-value bound is a dynamic count of the
// controller's permanents ("with mana value less than or equal to the number of
// lands you control", Beseech the Queen). The count subject and its "you
// control" suffix must stay on the ManaValueDynamicCount rider rather than
// folding into the searched card's own type or controller, and the amount must
// remain the singular one so the search finds exactly one card.
func TestParseSearchManaValueDynamicCountBound(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Search your library for a card with mana value less than or equal to the number of lands you control, reveal it, put it into your hand, then shuffle.",
		Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	effect := &document.Abilities[0].Sentences[0].Effects[0]
	if effect.Kind != EffectSearch {
		t.Fatalf("effect kind = %v, want EffectSearch", effect.Kind)
	}
	if effect.UnsupportedDetail != "" {
		t.Fatalf("unexpected unsupported detail %q", effect.UnsupportedDetail)
	}
	if !effect.Amount.Known || effect.Amount.Value != 1 {
		t.Fatalf("search amount = %+v, want the singular card", effect.Amount)
	}

	// The searched card is a bare "a card": no type, controller, or static
	// mana-value bound must leak onto it from the count subject.
	if effect.Selection.Kind != SelectionCard {
		t.Fatalf("searched selection kind = %v, want SelectionCard", effect.Selection.Kind)
	}
	if len(effect.Selection.RequiredTypesAny) != 0 {
		t.Fatalf("searched selection leaked required types %v", effect.Selection.RequiredTypesAny)
	}
	if effect.Selection.Controller != "" {
		t.Fatalf("searched selection leaked controller %q", effect.Selection.Controller)
	}
	if effect.Selection.MatchManaValue {
		t.Fatal("searched selection gained a static mana-value bound")
	}

	bound := effect.Selection.ManaValueDynamicCount
	if bound == nil {
		t.Fatal("searched selection has no dynamic-count mana-value bound")
	}
	if bound.DynamicKind != EffectDynamicAmountCount {
		t.Fatalf("bound dynamic kind = %v, want EffectDynamicAmountCount", bound.DynamicKind)
	}
	if bound.Multiplier != 1 {
		t.Fatalf("bound multiplier = %d, want 1", bound.Multiplier)
	}
	if bound.Selection.Kind != SelectionLand {
		t.Fatalf("count subject kind = %v, want SelectionLand", bound.Selection.Kind)
	}
	if bound.Selection.Controller != SelectionControllerYou {
		t.Fatalf("count subject controller = %v, want you", bound.Selection.Controller)
	}
}
