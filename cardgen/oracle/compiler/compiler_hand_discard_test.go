package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestCompileHandDiscardIsTextBlind(t *testing.T) {
	t.Parallel()
	document, diagnostics := parser.Parse(
		"Draw two cards, then discard two cards.",
		parser.Context{InstantOrSorcery: true, CardName: "Faithless Looting"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	discard := &document.Abilities[0].Sentences[0].Effects[1]
	discard.Text = "downstream must not inspect this"
	discard.Tokens = nil

	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 2 || effects[1].Kind != EffectDiscard ||
		!effects[1].Exact || !effects[1].HandDiscard.Present {
		t.Fatalf("compiled effects = %#v, want typed hand discard", effects)
	}
}

func TestCompileDiscardEntireHandIsTextBlind(t *testing.T) {
	t.Parallel()
	document, diagnostics := parser.Parse(
		"Each player discards their hand.",
		parser.Context{InstantOrSorcery: true, CardName: "Windfall"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	discard := &document.Abilities[0].Sentences[0].Effects[0]
	discard.Text = "downstream must not inspect this"
	discard.Tokens = nil

	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 1 || effects[0].Kind != EffectDiscard ||
		!effects[0].Exact || !effects[0].DiscardEntireHand {
		t.Fatalf("compiled effects = %#v, want typed discard entire hand", effects)
	}
}
