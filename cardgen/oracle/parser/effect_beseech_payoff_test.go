package parser

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// TestParseBargainSearchCastPayoffSequence proves the parser recognizes Beseech
// the Mirror's "Search your library for a card, exile it face down, then shuffle.
// If this spell was bargained, you may cast the exiled card without paying its
// mana cost if that spell's mana value is 4 or less. Put the exiled card into
// your hand if it wasn't cast this way." as an exact sequence carrying the
// mana-value bound.
func TestParseBargainSearchCastPayoffSequence(t *testing.T) {
	source := "Search your library for a card, exile it face down, then shuffle. " +
		"If this spell was bargained, you may cast the exiled card without paying its mana cost " +
		"if that spell's mana value is 4 or less. Put the exiled card into your hand if it wasn't cast this way."

	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})

	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v diagnostics = %#v, want one ability", document.Abilities, diagnostics)
	}
	sequence := document.Abilities[0].ExactSequence
	if sequence == nil || sequence.Kind != ExactSequenceBargainSearchCastPayoff {
		t.Fatalf("exact sequence = %#v, want bargain-search-cast-payoff sequence", sequence)
	}
	if sequence.MaxManaValue != 4 {
		t.Fatalf("sequence max mana value = %d, want 4", sequence.MaxManaValue)
	}
	if sequence.Span == (shared.Span{}) || sequence.Span != document.Abilities[0].BodySpan {
		t.Fatalf("sequence span = %#v body span = %#v, want exact body span", sequence.Span, document.Abilities[0].BodySpan)
	}
}
