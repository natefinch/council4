package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestCompileHandLibraryPutIsTextBlind(t *testing.T) {
	t.Parallel()
	source := "Draw three cards, then put two cards from your hand on top of your library in any order."
	document, diagnostics := parser.Parse(source, parser.Context{InstantOrSorcery: true, CardName: "Brainstorm"})
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	put := &document.Abilities[0].Sentences[0].Effects[1]
	put.Text = "downstream must not inspect this"
	put.Tokens = nil

	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 2 || effects[1].Kind != EffectPut ||
		!effects[1].Exact || !effects[1].HandLibraryPut.Present {
		t.Fatalf("compiled effects = %#v, want typed hand-library put", effects)
	}
}
