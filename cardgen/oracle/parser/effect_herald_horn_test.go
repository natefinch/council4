package parser

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func TestParseChosenTypeLibraryTopSequence(t *testing.T) {
	source := "At the beginning of your upkeep, look at the top card of your library. If it's a creature card of the chosen type, you may reveal it and put it into your hand."

	document, diagnostics := Parse(source, Context{})

	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v diagnostics = %#v, want one ability", document.Abilities, diagnostics)
	}
	sequence := document.Abilities[0].ExactSequence
	if sequence == nil || sequence.Kind != ExactSequenceChosenTypeLibraryTopToHand {
		t.Fatalf("exact sequence = %#v, want chosen-type library-top sequence", sequence)
	}
	if sequence.Span == (shared.Span{}) || sequence.Span != document.Abilities[0].BodySpan {
		t.Fatalf("sequence span = %#v body span = %#v, want exact body span", sequence.Span, document.Abilities[0].BodySpan)
	}
}
