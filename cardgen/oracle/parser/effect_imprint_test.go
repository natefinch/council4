package parser

import "testing"

const mimicVatOracle = "Imprint — Whenever a nontoken creature dies, you may exile that card. If you do, return each other card exiled with this artifact to its owner's graveyard.\n{3}, {T}: Create a token that's a copy of a card exiled with this artifact. It gains haste. Exile it at the beginning of the next end step."

func TestParseMimicVatExactSequences(t *testing.T) {
	document, diagnostics := Parse(mimicVatOracle, Context{CardName: "Mimic Vat"})
	if len(diagnostics) != 0 || len(document.Abilities) != 2 {
		t.Fatalf("abilities = %#v diagnostics = %#v, want two abilities", document.Abilities, diagnostics)
	}
	if sequence := document.Abilities[0].ExactSequence; sequence == nil ||
		sequence.Kind != ExactSequenceReplaceLinkedExiledCard {
		t.Fatalf("imprint sequence = %#v, want died-creature imprint", sequence)
	}
	if sequence := document.Abilities[1].ExactSequence; sequence == nil ||
		sequence.Kind != ExactSequenceLinkedExiledCopyToken {
		t.Fatalf("activation sequence = %#v, want temporary imprint copy", sequence)
	}
}

func TestParseMimicVatExactSequencesFailClosed(t *testing.T) {
	tests := []string{
		"Whenever a nontoken creature dies, you may exile that card. If you do, return each other card exiled with this artifact to its owner's hand.",
		"Create a token that's a copy of a permanent exiled with this artifact. It gains haste. Exile it at the beginning of the next end step.",
		"Create a token that's a copy of a card exiled with this artifact. It gains haste. Sacrifice it at the beginning of the next end step.",
	}
	for _, source := range tests {
		document, _ := Parse(source, Context{})
		for _, ability := range document.Abilities {
			if ability.ExactSequence != nil &&
				(ability.ExactSequence.Kind == ExactSequenceReplaceLinkedExiledCard ||
					ability.ExactSequence.Kind == ExactSequenceLinkedExiledCopyToken) {
				t.Fatalf("near miss recognized as Mimic Vat exact sequence: %q", source)
			}
		}
	}
}
