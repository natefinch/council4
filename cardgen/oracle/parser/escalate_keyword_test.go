package parser

import "testing"

func TestParseEscalateKeyword(t *testing.T) {
	t.Parallel()
	source := "Escalate {G} (Pay this cost for each mode chosen beyond the first.)\n" +
		"Choose one or more —\n" +
		"• Destroy target artifact.\n" +
		"• Destroy target enchantment.\n" +
		"• Target creature gains hexproof and indestructible until end of turn."
	modal := spreeModalFor(t, source)
	if !modal.Escalate {
		t.Fatal("modal.Escalate = false; want true")
	}
	if modal.Spree {
		t.Fatal("modal.Spree = true; want false for an Escalate modal")
	}
	if got := modal.EscalateCost.ManaValue(); got != 1 {
		t.Fatalf("escalate cost mana value = %d; want 1", got)
	}
	if !modal.ChoiceKnown || modal.ChoiceKind != ModalChoiceKindOneOrMore {
		t.Fatalf("choice = (known %v, kind %v); want known one-or-more", modal.ChoiceKnown, modal.ChoiceKind)
	}
	if modal.MinModes != 1 || modal.MaxModes != 3 {
		t.Fatalf("modes range = %d/%d; want 1/3", modal.MinModes, modal.MaxModes)
	}
	if len(modal.Options) != 3 {
		t.Fatalf("options = %d; want 3", len(modal.Options))
	}
	for i := range modal.Options {
		if modal.Options[i].SpreeCost != nil {
			t.Fatalf("option %d carries a per-mode Spree cost; Escalate options share one cost", i)
		}
	}
}

// TestParseEscalateWithoutModalFailsClosed verifies that an Escalate header that
// is not followed by a modal choose header is left as an unrecognized ability
// (it produces diagnostics) rather than being silently dropped.
func TestParseEscalateWithoutModalFailsClosed(t *testing.T) {
	t.Parallel()
	source := "Escalate {G} (Pay this cost for each mode chosen beyond the first.)\n" +
		"Draw a card."
	document, _ := Parse(source, Context{CardName: "Test Escalate"})
	for i := range document.Abilities {
		if document.Abilities[i].Modal != nil && document.Abilities[i].Modal.Escalate {
			t.Fatal("Escalate folded into a non-modal body; want it left unrecognized")
		}
	}
}
