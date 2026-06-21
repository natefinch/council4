package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// TestCompileChosenTypeLibraryTopIsTextBlind confirms the compiler carries the
// parser-recognized exact sequence as a typed kind without inspecting Oracle
// wording, and emits no unsupported diagnostic for the empty effect body.
func TestCompileChosenTypeLibraryTopIsTextBlind(t *testing.T) {
	t.Parallel()
	source := "At the beginning of your upkeep, look at the top card of your library. " +
		"If it's a creature card of the chosen type, you may reveal it and put it into your hand."
	document, diagnostics := parser.Parse(source, parser.Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	if document.Abilities[0].ExactSequence == nil {
		t.Fatal("parser did not recognize the chosen-type library-top sequence")
	}
	document.Abilities[0].Text = "downstream must not inspect this"

	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	compiledAbility := compilation.Abilities[0]
	if compiledAbility.ExactSequence != ExactSequenceChosenTypeLibraryTopToHand {
		t.Fatalf("exact sequence = %v, want chosen-type library-top", compiledAbility.ExactSequence)
	}
	content := compiledAbility.Content
	if len(content.Effects) != 0 {
		t.Fatalf("effects = %#v, want none; the body is carried as an exact sequence", content.Effects)
	}
}
