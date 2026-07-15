package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// TestCompileCarriesCorrelatedDistributiveFight proves the compiler carries the
// parser's CorrelatedDistributiveFight flag onto the compiled fight effect, so the
// executable backend can recognize the correlated group-fight sequence (Ezuri's
// Predation) and pair the created-token group with the counted-permanent group.
func TestCompileCarriesCorrelatedDistributiveFight(t *testing.T) {
	t.Parallel()
	document, diagnostics := parser.Parse(
		"For each creature your opponents control, create a 4/4 green Phyrexian Beast creature token. Each of those tokens fights a different one of those creatures.",
		parser.Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	var fight *CompiledEffect
	for i := range compilation.Abilities[0].Content.Effects {
		effect := &compilation.Abilities[0].Content.Effects[i]
		if effect.Kind == EffectFight {
			fight = effect
			break
		}
	}
	if fight == nil {
		t.Fatal("no compiled EffectFight found")
	}
	if !fight.CorrelatedDistributiveFight {
		t.Fatal("CorrelatedDistributiveFight = false, want true on the compiled fight effect")
	}
}
