package parser

import "testing"

func spreeModalFor(t *testing.T, source string) *Modal {
	t.Helper()
	document, diagnostics := Parse(source, Context{CardName: "Test Spree"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %+v; want none", diagnostics)
	}
	for i := range document.Abilities {
		if modal := document.Abilities[i].Modal; modal != nil {
			return modal
		}
	}
	t.Fatalf("no modal ability parsed from %q", source)
	return nil
}

func TestParseSpreeKeyword(t *testing.T) {
	t.Parallel()
	source := "Spree (Choose one or more additional costs.)\n" +
		"+ {1} — Search your library for a card, put it into your graveyard, then shuffle.\n" +
		"+ {2} — Return up to two creature cards with total mana value 4 or less from your graveyard to the battlefield."
	modal := spreeModalFor(t, source)
	if !modal.Spree {
		t.Fatal("modal.Spree = false; want true")
	}
	if !modal.ChoiceKnown || modal.ChoiceKind != ModalChoiceKindOneOrMore {
		t.Fatalf("choice = (known %v, kind %v); want known one-or-more", modal.ChoiceKnown, modal.ChoiceKind)
	}
	if modal.MinModes != 1 || modal.MaxModes != 2 {
		t.Fatalf("modes range = %d/%d; want 1/2", modal.MinModes, modal.MaxModes)
	}
	if len(modal.Options) != 2 {
		t.Fatalf("options = %d; want 2", len(modal.Options))
	}
	for i, want := range []int{1, 2} {
		clause := modal.Options[i].SpreeCost
		if clause == nil {
			t.Fatalf("option %d has no Spree cost clause", i)
		}
		if got := clause.Cost.ManaValue(); got != want {
			t.Fatalf("option %d cost mana value = %d; want %d", i, got, want)
		}
	}
}

func TestParseSpreeKeywordColoredCost(t *testing.T) {
	t.Parallel()
	source := "Spree (Choose one or more additional costs.)\n" +
		"+ {1} — All creatures lose all abilities until end of turn.\n" +
		"+ {3}{W}{W} — Destroy all creatures."
	modal := spreeModalFor(t, source)
	if len(modal.Options) != 2 || modal.MaxModes != 2 {
		t.Fatalf("modal options/max = %d/%d; want 2/2", len(modal.Options), modal.MaxModes)
	}
	second := modal.Options[1].SpreeCost
	if second == nil || second.Cost.ManaValue() != 5 {
		t.Fatalf("second option cost = %+v; want mana value 5", second)
	}
}
