package parser

import "testing"

func TestRecognizeModesUniquePerTurn(t *testing.T) {
	t.Parallel()

	source := "Whenever another creature you control enters, choose one that hasn't been chosen this turn —\n" +
		"• Put a +1/+1 counter on this creature.\n" +
		"• Create a tapped Treasure token.\n" +
		"• You gain 2 life."
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 || document.Abilities[0].Modal == nil {
		t.Fatalf("abilities = %#v, want one modal ability", document.Abilities)
	}
	modal := document.Abilities[0].Modal
	if !modal.ChoiceKnown || !modal.ModesUniquePerTurn ||
		modal.MinModes != 1 || modal.MaxModes != 1 {
		t.Fatalf("modal = %#v, want one mode unique per turn", modal)
	}
}

func TestModesUniquePerTurnHeaderFailsClosedOnExtraWords(t *testing.T) {
	t.Parallel()

	source := "Choose one that hasn't been chosen this turn if able —\n" +
		"• Draw a card.\n" +
		"• You gain 1 life."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if document.Abilities[0].Modal.ModesUniquePerTurn {
		t.Fatal("inexact modal header was marked unique per turn")
	}
}
