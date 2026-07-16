package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/counter"
)

func TestCompileOptionalCounterForEachPlayerIsTextBlind(t *testing.T) {
	t.Parallel()
	const oracle = "At the beginning of your end step, each player may put two +1/+1 counters on a creature they control. Goad each creature that had counters put on it this way. (Until your next turn, those creatures attack each combat if able and attack a player other than you if able.)"
	document, diagnostics := parser.Parse(oracle, parser.Context{CardName: "Agitator Ant"})
	if len(diagnostics) != 0 {
		t.Fatalf("parser diagnostics = %#v", diagnostics)
	}
	document.Abilities[0].Text = "compiler must not inspect Oracle text"
	for i := range document.Abilities[0].Sentences {
		document.Abilities[0].Sentences[i].Text = "opaque"
		document.Abilities[0].Sentences[i].Tokens = nil
	}

	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compiler diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 1 {
		t.Fatalf("effects = %#v, want one", effects)
	}
	effect := effects[0]
	if !effect.OptionalCounterForEachPlayer ||
		effect.Context != parser.EffectContextEachPlayer ||
		!effect.Amount.Known ||
		effect.Amount.Value != 2 ||
		!effect.CounterKindKnown ||
		effect.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("effect = %#v", effect)
	}
}
