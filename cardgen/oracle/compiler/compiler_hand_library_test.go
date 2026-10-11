package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestCompileHandLibraryPutIsTextBlind(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Draw three cards, then put two cards from your hand on top of your library in any order.",
		"Put X cards from your hand on the bottom of your library.",
		"Put two cards from your hand on top of your library in a random order.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := parser.Parse(source, parser.Context{InstantOrSorcery: true, CardName: "Test"})
			if len(diagnostics) != 0 {
				t.Fatalf("parse diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			put := &effects[len(effects)-1]
			payload, exact := put.HandLibraryPut, put.Exact
			put.Text = "downstream must not inspect this"
			put.Tokens = nil
			compilation, diagnostics := Compile(document, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("compile diagnostics = %#v", diagnostics)
			}
			compiled := compilation.Abilities[0].Content.Effects
			got := compiled[len(compiled)-1]
			if got.Kind != EffectPut || got.Exact != exact || got.HandLibraryPut != payload {
				t.Fatalf("compiled payload=%+v exact=%v, want %+v/%v", got.HandLibraryPut, got.Exact, payload, exact)
			}
			if put.Amount.VariableX != got.Amount.VariableX || put.Amount.Value != got.Amount.Value {
				t.Fatalf("compiled amount=%+v, want %+v", got.Amount, put.Amount)
			}
		})
	}
}
