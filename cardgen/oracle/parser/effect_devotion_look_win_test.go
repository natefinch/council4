package parser

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// TestParseDevotionLookWinSequence proves the parser recognizes Thassa's
// Oracle's "When this creature enters, look at the top X cards of your library,
// where X is your devotion to blue. Put up to one of them on top of your library
// and the rest on the bottom of your library in a random order. If X is greater
// than or equal to the number of cards in your library, you win the game." as an
// exact sequence carrying the devotion color, with the trailing reminder text
// excluded from matching.
func TestParseDevotionLookWinSequence(t *testing.T) {
	source := "When this creature enters, look at the top X cards of your library, " +
		"where X is your devotion to blue. Put up to one of them on top of your library " +
		"and the rest on the bottom of your library in a random order. " +
		"If X is greater than or equal to the number of cards in your library, you win the game. " +
		"(Each {U} in the mana costs of permanents you control counts toward your devotion to blue.)"

	document, diagnostics := Parse(source, Context{})

	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v diagnostics = %#v, want one ability", document.Abilities, diagnostics)
	}
	sequence := document.Abilities[0].ExactSequence
	if sequence == nil || sequence.Kind != ExactSequenceDevotionLookWin {
		t.Fatalf("exact sequence = %#v, want devotion-look-win sequence", sequence)
	}
	if sequence.DevotionColor != ColorBlue {
		t.Fatalf("sequence devotion color = %q, want ColorBlue", sequence.DevotionColor)
	}
	if sequence.Span == (shared.Span{}) || sequence.Span != document.Abilities[0].BodySpan {
		t.Fatalf("sequence span = %#v body span = %#v, want exact body span", sequence.Span, document.Abilities[0].BodySpan)
	}
}

// TestParseDevotionLookWinSequenceGenericColor proves the recognizer is generic
// over the devotion color rather than hard-coded to blue.
func TestParseDevotionLookWinSequenceGenericColor(t *testing.T) {
	source := "When this creature enters, look at the top X cards of your library, " +
		"where X is your devotion to red. Put up to one of them on top of your library " +
		"and the rest on the bottom of your library in a random order. " +
		"If X is greater than or equal to the number of cards in your library, you win the game."

	document, diagnostics := Parse(source, Context{})

	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v diagnostics = %#v, want one ability", document.Abilities, diagnostics)
	}
	sequence := document.Abilities[0].ExactSequence
	if sequence == nil || sequence.Kind != ExactSequenceDevotionLookWin {
		t.Fatalf("exact sequence = %#v, want devotion-look-win sequence", sequence)
	}
	if sequence.DevotionColor != ColorRed {
		t.Fatalf("sequence devotion color = %q, want ColorRed", sequence.DevotionColor)
	}
}
