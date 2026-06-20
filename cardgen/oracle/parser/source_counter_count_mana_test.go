package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
)

func TestParseSourceCounterCountManaTypedSyntax(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		text    string
		color   mana.Color
		counter counter.Kind
	}{
		{"this artifact", "Add {C} for each charge counter on this artifact.", mana.C, counter.Charge},
		{"this enchantment", "Add {B} for each charge counter on this enchantment.", mana.B, counter.Charge},
		{"pronoun it", "Add {B} for each charge counter on it.", mana.B, counter.Charge},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(tc.text, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
				t.Fatalf("abilities = %#v", document.Abilities)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v", effects)
			}
			effect := effects[0]
			if effect.Kind != EffectAddMana || !effect.Exact ||
				effect.Amount.DynamicKind != EffectDynamicAmountSourceCounterCount ||
				effect.Amount.DynamicForm != EffectDynamicAmountFormForEach ||
				effect.Amount.Multiplier != 1 ||
				effect.Amount.CounterKind != tc.counter {
				t.Fatalf("amount = %#v", effect.Amount)
			}
			if !effect.Mana.ColorsKnown || len(effect.Mana.Colors) != 1 ||
				effect.Mana.Colors[0] != tc.color || effect.Mana.Choice || effect.Mana.AnyColor {
				t.Fatalf("mana = %#v", effect.Mana)
			}
		})
	}
}

func TestParseSourceCounterCountSelfName(t *testing.T) {
	t.Parallel()
	// The card's own name names the source permanent just like a "this <type>"
	// marker (Black Market's "add {B} for each charge counter on Black Market").
	document, diagnostics := Parse(
		"Add {B} for each charge counter on Black Market.",
		Context{CardName: "Black Market"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Amount.DynamicKind != EffectDynamicAmountSourceCounterCount ||
		!effects[0].Mana.ColorsKnown || len(effects[0].Mana.Colors) != 1 || effects[0].Mana.Colors[0] != mana.B {
		t.Fatalf("effects = %#v", effects)
	}
}
