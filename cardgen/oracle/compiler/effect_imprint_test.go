package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestCompileMimicVatExactSequencesIsTextBlind(t *testing.T) {
	const oracle = "Imprint — Whenever a nontoken creature dies, you may exile that card. If you do, return each other card exiled with this artifact to its owner's graveyard.\n{3}, {T}: Create a token that's a copy of a card exiled with this artifact. It gains haste. Exile it at the beginning of the next end step."
	document, diagnostics := parser.Parse(oracle, parser.Context{CardName: "Mimic Vat"})
	if len(diagnostics) != 0 || len(document.Abilities) != 2 {
		t.Fatalf("parse diagnostics = %#v abilities = %#v", diagnostics, document.Abilities)
	}
	for i := range document.Abilities {
		document.Abilities[i].Text = "downstream must not inspect this"
	}

	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 || len(compilation.Abilities) != 2 {
		t.Fatalf("compile diagnostics = %#v abilities = %#v", diagnostics, compilation.Abilities)
	}
	if got := compilation.Abilities[0].ExactSequence; got != ExactSequenceReplaceLinkedExiledCard {
		t.Fatalf("trigger exact sequence = %v, want imprint died creature", got)
	}
	if got := compilation.Abilities[1].ExactSequence; got != ExactSequenceLinkedExiledCopyToken {
		t.Fatalf("activation exact sequence = %v, want temporary imprint copy", got)
	}
}
