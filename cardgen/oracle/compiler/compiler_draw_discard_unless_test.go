package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestCompileDrawThenDiscardUnlessIsTextBlind confirms the compiler carries the
// parser-recognized "Draw N cards. Then discard M cards unless you discard a
// <type> card." sequence as a typed kind with its counts and exempt types,
// without inspecting Oracle wording and without emitting an unsupported
// diagnostic for the empty effect body.
func TestCompileDrawThenDiscardUnlessIsTextBlind(t *testing.T) {
	t.Parallel()
	document, diagnostics := parser.Parse(
		"Draw three cards. Then discard two cards unless you discard an artifact card.",
		parser.Context{InstantOrSorcery: true, CardName: "Thirst for Knowledge"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	if document.Abilities[0].ExactSequence == nil {
		t.Fatal("parser did not recognize the draw-then-discard-unless sequence")
	}
	document.Abilities[0].Text = "downstream must not inspect this"

	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.ExactSequence != ExactSequenceDrawThenDiscardUnlessType {
		t.Fatalf("exact sequence = %v, want draw-then-discard-unless", ability.ExactSequence)
	}
	if ability.ExactSequenceDrawCount != 3 || ability.ExactSequenceDiscardCount != 2 {
		t.Fatalf("counts draw=%d discard=%d, want 3 and 2",
			ability.ExactSequenceDrawCount, ability.ExactSequenceDiscardCount)
	}
	if len(ability.ExactSequenceLookAtTopTypes) != 1 ||
		ability.ExactSequenceLookAtTopTypes[0] != types.Artifact {
		t.Fatalf("exempt types = %#v, want [artifact]", ability.ExactSequenceLookAtTopTypes)
	}
	if len(ability.Content.Effects) != 0 {
		t.Fatalf("effects = %#v, want none; the body is carried as an exact sequence", ability.Content.Effects)
	}
}

// TestCompileDrawThenDiscardUnlessDisjunction confirms the multi-type exempt
// disjunction ("an instant or sorcery card") is carried as both typed values.
func TestCompileDrawThenDiscardUnlessDisjunction(t *testing.T) {
	t.Parallel()
	document, diagnostics := parser.Parse(
		"Draw four cards. Then discard two cards unless you discard an instant or sorcery card.",
		parser.Context{InstantOrSorcery: true, CardName: "Practical Research"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.ExactSequence != ExactSequenceDrawThenDiscardUnlessType {
		t.Fatalf("exact sequence = %v, want draw-then-discard-unless", ability.ExactSequence)
	}
	if ability.ExactSequenceDrawCount != 4 {
		t.Fatalf("draw count = %d, want 4", ability.ExactSequenceDrawCount)
	}
	want := []types.Card{types.Instant, types.Sorcery}
	if len(ability.ExactSequenceLookAtTopTypes) != len(want) {
		t.Fatalf("exempt types = %#v, want %v", ability.ExactSequenceLookAtTopTypes, want)
	}
	for i, ct := range want {
		if ability.ExactSequenceLookAtTopTypes[i] != ct {
			t.Fatalf("exempt type[%d] = %v, want %v", i, ability.ExactSequenceLookAtTopTypes[i], ct)
		}
	}
}
