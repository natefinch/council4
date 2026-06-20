package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestCompileLibraryTopReorderIsTextBlind(t *testing.T) {
	t.Parallel()
	source := "Look at the top three cards of your library, then put them back in any order."
	document, diagnostics := parser.Parse(source, parser.Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	effect := &document.Abilities[0].Sentences[0].Effects[0]
	effect.Text = "downstream must not inspect this"
	effect.Tokens = nil

	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 1 ||
		effects[0].Kind != EffectReorderLibraryTop ||
		!effects[0].Exact ||
		!effects[0].Amount.Known ||
		effects[0].Amount.Value != 3 {
		t.Fatalf("compiled effects = %#v, want typed reorder of three", effects)
	}
}
