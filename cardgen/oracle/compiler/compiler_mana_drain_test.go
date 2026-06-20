package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

func TestCompileManaDrainUsesTypedDelayedMana(t *testing.T) {
	t.Parallel()
	document, diagnostics := parser.Parse(
		"Counter target spell. At the beginning of your next main phase, add an amount of {C} equal to that spell's mana value.",
		parser.Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	var effects []CompiledEffect
	for _, ability := range compilation.Abilities {
		effects = append(effects, ability.Content.Effects...)
	}
	if len(effects) != 2 {
		t.Fatalf("effects = %#v, want two", effects)
	}
	effect := effects[1]
	if effect.Kind != EffectAddMana ||
		!effect.Exact ||
		effect.DelayedTiming != game.DelayedAtBeginningOfNextMainPhase ||
		!effect.Mana.DynamicColorless ||
		effect.Amount.DynamicKind != DynamicAmountSourceManaValue ||
		effect.Amount.DynamicForm != DynamicAmountEqual ||
		len(effect.References) != 1 ||
		effect.References[0].Binding != ReferenceBindingTarget ||
		effect.References[0].Occurrence != 0 {
		t.Fatalf("compiled delayed mana = %#v", effect)
	}
}
