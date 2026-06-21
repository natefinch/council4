package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// TestCompileBottomHandThenDrawIsTextBlind confirms the compiler carries the
// parser-recognized bottom-hand-then-draw sequence and its typed parameters
// without inspecting Oracle wording, and emits no unsupported diagnostic for the
// empty effect body.
func TestCompileBottomHandThenDrawIsTextBlind(t *testing.T) {
	t.Parallel()
	source := "Put any number of cards from your hand on the bottom of your library, then draw that many cards plus one."
	document, diagnostics := parser.Parse(source, parser.Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	if document.Abilities[0].ExactSequence == nil {
		t.Fatal("parser did not recognize the bottom-hand-then-draw sequence")
	}
	document.Abilities[0].Text = "downstream must not inspect this"

	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	compiledAbility := compilation.Abilities[0]
	if compiledAbility.ExactSequence != ExactSequenceBottomHandThenDraw {
		t.Fatalf("exact sequence = %v, want bottom-hand-then-draw", compiledAbility.ExactSequence)
	}
	if !compiledAbility.ExactSequenceBottom {
		t.Fatalf("exact sequence bottom = %v, want true", compiledAbility.ExactSequenceBottom)
	}
	if compiledAbility.ExactSequenceDrawOffset != 1 {
		t.Fatalf("exact sequence draw offset = %d, want 1", compiledAbility.ExactSequenceDrawOffset)
	}
	if len(compiledAbility.Content.Effects) != 0 {
		t.Fatalf("effects = %#v, want none; the body is carried as an exact sequence", compiledAbility.Content.Effects)
	}
}
