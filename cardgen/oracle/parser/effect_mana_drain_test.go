package parser

import "testing"

const manaDrainOracle = "Counter target spell. At the beginning of your next main phase, add an amount of {C} equal to that spell's mana value."

func TestParseManaDrainTypedDelayedMana(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(manaDrainOracle, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	var effects []EffectSyntax
	for _, ability := range document.Abilities {
		for _, sentence := range ability.Sentences {
			effects = append(effects, sentence.Effects...)
		}
	}
	if len(effects) != 2 {
		t.Fatalf("effects = %#v, want counter and delayed mana", effects)
	}
	effect := effects[1]
	if effect.Kind != EffectAddMana ||
		!effect.Exact ||
		effect.DelayedTiming != DelayedTimingNextMain ||
		!effect.Mana.DynamicColorless ||
		effect.Amount.DynamicKind != EffectDynamicAmountSourceManaValue ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormEqual ||
		effect.Amount.Multiplier != 1 ||
		len(effect.References) != 1 {
		t.Fatalf("delayed mana effect = %#v", effect)
	}
}

func TestParseManaDrainNearMissesFailClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Counter target spell. At the beginning of the next main phase, add an amount of {C} equal to that spell's mana value.",
		"Counter target spell. At the beginning of your next upkeep, add an amount of {C} equal to that spell's mana value.",
		"Counter target spell. At the beginning of your next main phase, add an amount of {U} equal to that spell's mana value.",
		"Counter target spell. At the beginning of your next main phase, add an amount of {C} equal to its mana value.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(source, Context{InstantOrSorcery: true})
			var exact bool
			for _, ability := range document.Abilities {
				for _, sentence := range ability.Sentences {
					for _, effect := range sentence.Effects {
						exact = exact || effect.Kind == EffectAddMana && effect.Exact && effect.Mana.DynamicColorless
					}
				}
			}
			if exact {
				t.Fatal("near miss parsed as exact Mana Drain rider")
			}
		})
	}
}
