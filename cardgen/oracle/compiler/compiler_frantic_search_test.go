package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestCompileFranticSearchUntapIsTextBlind(t *testing.T) {
	t.Parallel()
	document, diagnostics := parser.Parse(
		"Draw two cards, then discard two cards. Untap up to three lands.",
		parser.Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	untap := &document.Abilities[0].Sentences[1].Effects[0]
	untap.Text = "downstream must not inspect this"
	untap.Tokens = nil
	untap.Selection.Text = "or this"

	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 3 ||
		effects[2].Kind != EffectUntap ||
		!effects[2].Exact ||
		!effects[2].Amount.RangeKnown ||
		effects[2].Amount.Minimum != 0 ||
		effects[2].Amount.Maximum != 3 ||
		effects[2].Selector.Kind != SelectorLand {
		t.Fatalf("compiled effects = %#v, want typed bounded land untap", effects)
	}
}
