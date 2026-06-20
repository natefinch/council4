package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestParseControlledCountManaTypedSyntax(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		text  string
		color mana.Color
	}{
		{"swamp", "Add {B} for each Swamp you control.", mana.B},
		{"creature", "Add {G} for each creature you control.", mana.G},
		{"enchantment", "Add {W} for each enchantment you control.", mana.W},
		{"artifact", "Add {U} for each artifact you control.", mana.U},
		{"basic swamp", "Add {B} for each basic Swamp you control.", mana.B},
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
				effect.Amount.DynamicKind != EffectDynamicAmountCount ||
				effect.Amount.DynamicForm != EffectDynamicAmountFormForEach ||
				effect.Amount.Multiplier != 1 ||
				effect.Amount.Selection == nil ||
				effect.Amount.Selection.Zone != zone.None ||
				effect.Amount.Selection.Controller != SelectionControllerYou {
				t.Fatalf("amount = %#v", effect.Amount)
			}
			if !effect.Mana.ColorsKnown || len(effect.Mana.Colors) != 1 ||
				effect.Mana.Colors[0] != tc.color || effect.Mana.Choice || effect.Mana.AnyColor {
				t.Fatalf("mana = %#v", effect.Mana)
			}
		})
	}
}

func TestParseControlledCountManaFailsClosed(t *testing.T) {
	t.Parallel()
	// A multi-symbol or any-color produced output is not modeled by the
	// single-color count recognizer, so the produced color must stay unset.
	variants := []string{
		"Add {G}{G} for each creature you control.",
		"Add one mana of any color for each creature you control.",
	}
	for _, source := range variants {
		document, _ := Parse(source, Context{})
		if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
			continue
		}
		effects := document.Abilities[0].Sentences[0].Effects
		if len(effects) != 1 {
			continue
		}
		if effects[0].Mana.ColorsKnown && len(effects[0].Mana.Colors) == 1 {
			t.Fatalf("variant unexpectedly recognized a single produced color:\n%s", source)
		}
	}
}
