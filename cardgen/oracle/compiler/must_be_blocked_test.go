package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestCompileTargetMustBeBlockedThisCombat(t *testing.T) {
	t.Parallel()
	document, parseDiagnostics := parser.Parse(
		"Target creature must be blocked this combat if able.",
		parser.Context{InstantOrSorcery: true},
	)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", parseDiagnostics)
	}
	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	if len(compilation.Abilities) != 1 {
		t.Fatalf("abilities = %#v, want one", compilation.Abilities)
	}
	content := compilation.Abilities[0].Content
	if len(content.Targets) != 1 ||
		!content.Targets[0].Exact ||
		content.Targets[0].Selector.Kind != SelectorCreature ||
		len(content.Effects) != 1 ||
		!content.Effects[0].Exact ||
		content.Effects[0].Kind != EffectMustBeBlocked {
		t.Fatalf("content = %#v, want exact creature target must-block effect", content)
	}
}
