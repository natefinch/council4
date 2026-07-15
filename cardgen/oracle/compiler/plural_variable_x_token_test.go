package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// TestCompilePluralVariableXTokenCount proves the compiler carries the explicit
// plural token count (three) on CompiledEffect.TokenCount while the "where X is
// <dynamic>" size dynamic rides on Amount, so lowering can emit a fixed count of
// dynamically-sized tokens rather than a single one.
func TestCompilePluralVariableXTokenCount(t *testing.T) {
	t.Parallel()
	source := "Create three tapped X/X green Treefolk creature tokens, where X is the amount of life you gained this turn."
	document, diagnostics := parser.Parse(source, parser.Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) == 0 || effects[0].Kind != EffectCreate {
		t.Fatalf("compiled effects = %#v, want a leading create", effects)
	}
	create := effects[0]
	if !create.TokenCount.Known || create.TokenCount.Value != 3 {
		t.Fatalf("token count = %+v, want a known fixed 3", create.TokenCount)
	}
	if !create.TokenPTVariableX {
		t.Fatalf("create = %+v, want variable X/X size", create)
	}
}
