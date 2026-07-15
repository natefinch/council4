package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// TestCompileKickerScaledTargetPropagates verifies the parser's folded
// KickerScaledCount target survives compilation onto the aggregated ability
// content, so cardgen can lower the kicker-scaled "each of them" damage spell
// (Comet Storm).
func TestCompileKickerScaledTargetPropagates(t *testing.T) {
	t.Parallel()
	document, parseDiagnostics := parser.Parse(
		"Choose any target, then choose another target for each time this spell was kicked. Comet Storm deals X damage to each of them.",
		parser.Context{CardName: "Comet Storm", InstantOrSorcery: true},
	)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", parseDiagnostics)
	}
	compilation, compileDiagnostics := Compile(document, Context{})
	if len(compileDiagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", compileDiagnostics)
	}
	content := compilation.Abilities[0].Content
	if len(content.Targets) != 1 {
		t.Fatalf("targets = %d, want 1 folded target", len(content.Targets))
	}
	target := content.Targets[0]
	if !target.KickerScaledCount {
		t.Fatal("target.KickerScaledCount = false, want true")
	}
	if target.Selector.Kind != SelectorAny {
		t.Fatalf("target.Selector.Kind = %v, want SelectorAny", target.Selector.Kind)
	}
	if len(content.Effects) != 1 || content.Effects[0].Kind != EffectDealDamage {
		t.Fatalf("effects = %#v, want a single deal-damage effect", content.Effects)
	}
	if !content.Effects[0].Exact {
		t.Fatal("damage effect Exact = false, want true for 'deals X damage to each of them'")
	}
	if !content.Effects[0].Amount.VariableX {
		t.Fatalf("damage amount = %#v, want VariableX", content.Effects[0].Amount)
	}
}
