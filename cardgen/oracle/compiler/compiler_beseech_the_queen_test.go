package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// TestCompileSearchManaValueDynamicCountBound proves the compiler preserves a
// search's dynamic-count mana-value bound ("with mana value less than or equal
// to the number of lands you control", Beseech the Queen) on the searched-card
// selector, compiling the count subject through the same typed-amount path as
// every other permanent count. It also confirms the compiler is text-blind: it
// reads the parsed structure, not the printed effect or selection text.
func TestCompileSearchManaValueDynamicCountBound(t *testing.T) {
	t.Parallel()
	document, diagnostics := parser.Parse(
		"Search your library for a card with mana value less than or equal to the number of lands you control, reveal it, put it into your hand, then shuffle.",
		parser.Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	search := &document.Abilities[0].Sentences[0].Effects[0]
	search.Text = "downstream must not inspect this"
	search.Tokens = nil
	search.Selection.Text = "or this"

	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) == 0 || effects[0].Kind != EffectSearch {
		t.Fatalf("compiled effects = %#v, want a leading search", effects)
	}
	selector := effects[0].Selector
	bound := selector.ManaValueDynamicCount
	if bound == nil {
		t.Fatal("compiled search selector has no dynamic-count mana-value bound")
	}
	if bound.DynamicKind != DynamicAmountCount {
		t.Fatalf("bound dynamic kind = %v, want DynamicAmountCount", bound.DynamicKind)
	}
	if bound.Multiplier != 1 {
		t.Fatalf("bound multiplier = %d, want 1", bound.Multiplier)
	}
	subject := bound.Selector()
	if subject.Kind != SelectorLand {
		t.Fatalf("count subject kind = %v, want SelectorLand", subject.Kind)
	}
	if subject.Controller != ControllerYou {
		t.Fatalf("count subject controller = %v, want you", subject.Controller)
	}
}
