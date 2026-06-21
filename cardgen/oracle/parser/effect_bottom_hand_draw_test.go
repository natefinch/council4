package parser

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// TestParseBottomHandThenDrawSequence proves the parser recognizes Valakut
// Awakening's "put any number of cards from your hand on the bottom of your
// library, then draw that many cards plus one." as an exact sequence carrying
// the library end and the draw offset.
func TestParseBottomHandThenDrawSequence(t *testing.T) {
	source := "Put any number of cards from your hand on the bottom of your library, then draw that many cards plus one."

	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})

	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v diagnostics = %#v, want one ability", document.Abilities, diagnostics)
	}
	sequence := document.Abilities[0].ExactSequence
	if sequence == nil || sequence.Kind != ExactSequenceBottomHandThenDraw {
		t.Fatalf("exact sequence = %#v, want bottom-hand-then-draw sequence", sequence)
	}
	if !sequence.Bottom {
		t.Fatalf("sequence bottom = %v, want true", sequence.Bottom)
	}
	if sequence.DrawOffset != 1 {
		t.Fatalf("sequence draw offset = %d, want 1", sequence.DrawOffset)
	}
	if sequence.Span == (shared.Span{}) || sequence.Span != document.Abilities[0].BodySpan {
		t.Fatalf("sequence span = %#v body span = %#v, want exact body span", sequence.Span, document.Abilities[0].BodySpan)
	}
}

// TestParseTopHandThenDrawNoOffset proves the recognizer also captures the
// top-of-library end and a zero offset ("draw that many cards" with no "plus").
func TestParseTopHandThenDrawNoOffset(t *testing.T) {
	source := "Put any number of cards from your hand on top of your library, then draw that many cards."

	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})

	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v diagnostics = %#v, want one ability", document.Abilities, diagnostics)
	}
	sequence := document.Abilities[0].ExactSequence
	if sequence == nil || sequence.Kind != ExactSequenceBottomHandThenDraw {
		t.Fatalf("exact sequence = %#v, want bottom-hand-then-draw sequence", sequence)
	}
	if sequence.Bottom {
		t.Fatalf("sequence bottom = %v, want false", sequence.Bottom)
	}
	if sequence.DrawOffset != 0 {
		t.Fatalf("sequence draw offset = %d, want 0", sequence.DrawOffset)
	}
}
