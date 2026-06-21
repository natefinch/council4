package parser

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func TestParseLabeledChooseOneOrMoreTrigger(t *testing.T) {
	t.Parallel()

	source := "At the beginning of your first main phase, choose one or more —\n" +
		"• Sell Contraband — Create a Treasure token. You lose 1 life.\n" +
		"• Buy Information — Draw a card. You lose 2 life.\n" +
		"• Hire a Mercenary — Create a 3/2 colorless Shapeshifter creature token with changeling. You lose 3 life."
	document, diagnostics := Parse(source, Context{CardName: "Test Broker"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v, want one", document.Abilities)
	}
	ability := document.Abilities[0]
	if ability.Kind != AbilityTriggered || ability.Modal == nil {
		t.Fatalf("ability = %#v, want triggered modal ability", ability)
	}
	if !ability.Modal.ChoiceKnown || ability.Modal.ChoiceKind != ModalChoiceKindOneOrMore ||
		ability.Modal.MinModes != 1 || ability.Modal.MaxModes != 3 {
		t.Fatalf("choice = %d..%d kind=%v known=%v, want typed one-or-more 1..3",
			ability.Modal.MinModes, ability.Modal.MaxModes, ability.Modal.ChoiceKind, ability.Modal.ChoiceKnown)
	}
	wantLabels := []ModeLabelKind{
		ModeLabelSellContraband,
		ModeLabelBuyInformation,
		ModeLabelHireMercenary,
	}
	if len(ability.Modal.Options) != len(wantLabels) {
		t.Fatalf("options = %#v, want %d", ability.Modal.Options, len(wantLabels))
	}
	for i, want := range wantLabels {
		mode := ability.Modal.Options[i]
		if mode.Label == nil || mode.Label.Kind != want {
			t.Fatalf("mode %d label = %#v, want %v", i, mode.Label, want)
		}
		if got := sharedText(source, mode.Label.Span); got != mode.Label.Text {
			t.Fatalf("mode %d label span text = %q, want %q", i, got, mode.Label.Text)
		}
		if got := sharedText(source, mode.Label.SeparatorSpan); got != "—" {
			t.Fatalf("mode %d separator span text = %q, want em dash", i, got)
		}
		if mode.Body.Text == "" || sharedText(source, mode.Body.Span) != mode.Body.Text {
			t.Fatalf("mode %d body = %#v, want exact source-backed body", i, mode.Body)
		}
	}
}

func sharedText(source string, span shared.Span) string {
	return shared.SliceSpan(source, span)
}

func TestParseChooseUpToOneTrigger(t *testing.T) {
	t.Parallel()

	source := "Whenever you cast a spell, choose up to one —\n" +
		"• Return target spell you don't control to its owner's hand.\n" +
		"• Return target nonland permanent to its owner's hand."
	document, diagnostics := Parse(source, Context{CardName: "Hullbreaker Horror"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v, want one", document.Abilities)
	}
	ability := document.Abilities[0]
	if ability.Kind != AbilityTriggered || ability.Modal == nil {
		t.Fatalf("ability = %#v, want triggered modal ability", ability)
	}
	// "choose up to one —" is an optional choice: minimum zero modes, maximum one.
	if !ability.Modal.ChoiceKnown || ability.Modal.MinModes != 0 || ability.Modal.MaxModes != 1 {
		t.Fatalf("choice = %d..%d known=%v, want optional 0..1",
			ability.Modal.MinModes, ability.Modal.MaxModes, ability.Modal.ChoiceKnown)
	}
	if len(ability.Modal.Options) != 2 {
		t.Fatalf("options = %#v, want two", ability.Modal.Options)
	}
}
