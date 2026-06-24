package parser

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// TestParseDiscardHandThenDrawSequence proves the parser recognizes Decaying
// Time Loop's whole-hand discard-then-draw-that-many body as an exact sequence,
// for both the verbose and terse hand wordings.
func TestParseDiscardHandThenDrawSequence(t *testing.T) {
	for _, source := range []string{
		"Discard all the cards in your hand, then draw that many cards.",
		"Discard your hand, then draw that many cards.",
	} {
		document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 || len(document.Abilities) != 1 {
			t.Fatalf("Parse(%q) abilities = %#v diagnostics = %#v", source, document.Abilities, diagnostics)
		}
		sequence := document.Abilities[0].ExactSequence
		if sequence == nil || sequence.Kind != ExactSequenceDiscardHandThenDraw {
			t.Fatalf("Parse(%q) exact sequence = %#v, want discard-hand-then-draw", source, sequence)
		}
		if sequence.Span == (shared.Span{}) || sequence.Span != document.Abilities[0].BodySpan {
			t.Fatalf("Parse(%q) sequence span = %#v body span = %#v", source, sequence.Span, document.Abilities[0].BodySpan)
		}
	}
}

// TestParseDiscardHandThenDrawFailsClosed proves the recognizer fires only on
// the exact whole-hand discard then "that many" draw; a fixed draw count, a
// trailing clause, or a partial-hand discard leaves the body unmatched.
func TestParseDiscardHandThenDrawFailsClosed(t *testing.T) {
	for _, source := range []string{
		"Discard your hand, then draw three cards.",
		"Discard a card, then draw that many cards.",
		"Discard all the cards in your hand, then draw that many cards. You gain 1 life.",
	} {
		document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 || len(document.Abilities) == 0 {
			continue
		}
		sequence := document.Abilities[0].ExactSequence
		if sequence != nil && sequence.Kind == ExactSequenceDiscardHandThenDraw {
			t.Errorf("Parse(%q) matched discard-hand-then-draw, want no match", source)
		}
	}
}
