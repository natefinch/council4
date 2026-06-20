package compiler

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/mana"
)

// TestCompileControlledCountManaCarriesTypedFields verifies that the compiler,
// without inspecting Oracle text, carries the parser's typed produced color and
// dynamic battlefield count for an "Add <mana> for each <permanent> you control"
// add-mana effect.
func TestCompileControlledCountManaCarriesTypedFields(t *testing.T) {
	t.Parallel()
	cases := []struct {
		source string
		color  mana.Color
	}{
		{"{T}: Add {B} for each Swamp you control.", mana.B},
		{"{T}: Add {G} for each creature you control.", mana.G},
		{"{T}: Add {W} for each enchantment you control.", mana.W},
	}
	for _, tc := range cases {
		compilation, diagnostics := compileSource(tc.source, pipelineContext{})
		if len(diagnostics) != 0 {
			t.Fatalf("diagnostics = %#v", diagnostics)
		}
		effect := compilation.Abilities[0].Content.Effects[0]
		if effect.Kind != EffectAddMana ||
			effect.Amount.DynamicKind != DynamicAmountCount ||
			effect.Amount.DynamicForm != DynamicAmountForEach {
			t.Fatalf("effect amount = %#v", effect.Amount)
		}
		if !effect.Mana.ColorsKnown || len(effect.Mana.Colors) != 1 ||
			effect.Mana.Colors[0] != tc.color {
			t.Fatalf("effect mana = %#v", effect.Mana)
		}
	}
}
