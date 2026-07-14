package parser

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

const braidsOracleText = "At the beginning of your end step, you may sacrifice an artifact, " +
	"creature, enchantment, land, or planeswalker. If you do, each opponent may " +
	"sacrifice a permanent of their choice that shares a card type with it. For each " +
	"opponent who doesn't, that player loses 2 life and you draw a card."

// TestParseSharedTypeSacrificePunisherSequence proves the parser recognizes
// Braids, Arisen Nightmare's resolving body as the shared-type-sacrifice
// punisher exact sequence spanning the whole body, so the compiler and lowerer
// can turn it into the generic optional-sacrifice-then-punisher primitives.
func TestParseSharedTypeSacrificePunisherSequence(t *testing.T) {
	document, diagnostics := Parse(braidsOracleText, Context{})

	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v diagnostics = %#v, want one ability", document.Abilities, diagnostics)
	}
	ability := document.Abilities[0]
	sequence := ability.ExactSequence
	if sequence == nil || sequence.Kind != ExactSequenceSharedTypeSacrificePunisher {
		t.Fatalf("exact sequence = %#v, want shared-type-sacrifice punisher sequence", sequence)
	}
	if sequence.Span == (shared.Span{}) || sequence.Span != ability.BodySpan {
		t.Fatalf("sequence span = %#v body span = %#v, want exact body span", sequence.Span, ability.BodySpan)
	}
}

// TestParseSharedTypeSacrificePunisherFailsClosed proves the recognizer is
// verbatim and fails closed: any deviation from the exact body — a different
// life amount or a dropped shared-card-type restriction — is not accepted as the
// exact sequence, so no card with altered wording is silently lowered as Braids.
func TestParseSharedTypeSacrificePunisherFailsClosed(t *testing.T) {
	cases := map[string]string{
		"different life amount": "At the beginning of your end step, you may sacrifice an artifact, " +
			"creature, enchantment, land, or planeswalker. If you do, each opponent may " +
			"sacrifice a permanent of their choice that shares a card type with it. For each " +
			"opponent who doesn't, that player loses 3 life and you draw a card.",
		"dropped shared-type restriction": "At the beginning of your end step, you may sacrifice an artifact, " +
			"creature, enchantment, land, or planeswalker. If you do, each opponent may " +
			"sacrifice a permanent of their choice. For each " +
			"opponent who doesn't, that player loses 2 life and you draw a card.",
		"missing controller draw": "At the beginning of your end step, you may sacrifice an artifact, " +
			"creature, enchantment, land, or planeswalker. If you do, each opponent may " +
			"sacrifice a permanent of their choice that shares a card type with it. For each " +
			"opponent who doesn't, that player loses 2 life.",
	}
	for name, source := range cases {
		t.Run(name, func(t *testing.T) {
			document, _ := Parse(source, Context{})
			for _, ability := range document.Abilities {
				if ability.ExactSequence != nil &&
					ability.ExactSequence.Kind == ExactSequenceSharedTypeSacrificePunisher {
					t.Fatal("altered wording was recognized as the shared-type-sacrifice punisher sequence, want fail-closed")
				}
			}
		})
	}
}
