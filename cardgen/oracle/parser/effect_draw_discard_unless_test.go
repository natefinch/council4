package parser

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// TestParseDrawThenDiscardUnlessSequence proves the parser recognizes the
// Thirst for Knowledge family body "Draw N cards. Then discard M cards unless
// you discard a <type> card." as an exact sequence, recording the counts and
// exempt card types for both the single-type and disjunctive forms, and for
// both the "a" and "an" article wordings.
func TestParseDrawThenDiscardUnlessSequence(t *testing.T) {
	cases := []struct {
		source     string
		drawCount  int
		discard    int
		exemptKind []CardType
	}{
		{"Draw three cards. Then discard two cards unless you discard a creature card.", 3, 2, []CardType{CardTypeCreature}},
		{"Draw three cards. Then discard two cards unless you discard an artifact card.", 3, 2, []CardType{CardTypeArtifact}},
		{"Draw four cards. Then discard two cards unless you discard an instant or sorcery card.", 4, 2, []CardType{CardTypeInstant, CardTypeSorcery}},
	}
	for _, tc := range cases {
		document, diagnostics := Parse(tc.source, Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 || len(document.Abilities) != 1 {
			t.Fatalf("Parse(%q) abilities = %#v diagnostics = %#v", tc.source, document.Abilities, diagnostics)
		}
		sequence := document.Abilities[0].ExactSequence
		if sequence == nil || sequence.Kind != ExactSequenceDrawThenDiscardUnlessType {
			t.Fatalf("Parse(%q) exact sequence = %#v, want draw-then-discard-unless", tc.source, sequence)
		}
		if sequence.DrawCount != tc.drawCount || sequence.DiscardCount != tc.discard {
			t.Fatalf("Parse(%q) draw=%d discard=%d, want %d and %d", tc.source, sequence.DrawCount, sequence.DiscardCount, tc.drawCount, tc.discard)
		}
		if len(sequence.LookAtTopCardTypes) != len(tc.exemptKind) {
			t.Fatalf("Parse(%q) exempt types = %#v, want %#v", tc.source, sequence.LookAtTopCardTypes, tc.exemptKind)
		}
		for i, want := range tc.exemptKind {
			if sequence.LookAtTopCardTypes[i] != want {
				t.Fatalf("Parse(%q) exempt[%d] = %v, want %v", tc.source, i, sequence.LookAtTopCardTypes[i], want)
			}
		}
		if sequence.Span == (shared.Span{}) || sequence.Span != document.Abilities[0].BodySpan {
			t.Fatalf("Parse(%q) sequence span = %#v body span = %#v", tc.source, sequence.Span, document.Abilities[0].BodySpan)
		}
	}
}

// TestParseDrawThenDiscardUnlessFailsClosed proves the recognizer fires only on
// the exact draw-then-discard-unless body; a missing exempt type, no rider, or
// trailing text leaves the body unmatched.
func TestParseDrawThenDiscardUnlessFailsClosed(t *testing.T) {
	for _, source := range []string{
		"Draw three cards. Then discard two cards.",
		"Draw three cards. Then discard two cards unless you discard a card.",
		"Draw three cards. Then discard two cards unless you discard a creature card. You gain 1 life.",
	} {
		document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 || len(document.Abilities) == 0 {
			continue
		}
		sequence := document.Abilities[0].ExactSequence
		if sequence != nil && sequence.Kind == ExactSequenceDrawThenDiscardUnlessType {
			t.Errorf("Parse(%q) matched draw-then-discard-unless, want no match", source)
		}
	}
}
