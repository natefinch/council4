package parser

import "testing"

const canoptekWraithOracle = "Wraith Form — This creature can't be blocked.\nTransdimensional Scout — When this creature deals combat damage to a player, you may pay {3} and sacrifice it. If you do, choose a land you control. Then search your library for up to two basic land cards which have the same name as the chosen land, put them onto the battlefield tapped, then shuffle."

func TestParseCanoptekWraithCompositeChoiceSearch(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(canoptekWraithOracle, Context{CardName: "Canoptek Wraith"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 2 {
		t.Fatalf("abilities = %#v, want two ability-word abilities", document.Abilities)
	}
	var effects []EffectSyntax
	for _, sentence := range document.Abilities[1].Sentences {
		effects = append(effects, sentence.Effects...)
	}
	if len(effects) != 5 {
		t.Fatalf("effects = %#v, want payment prelude, choice, search, put, shuffle", effects)
	}
	if effects[0].Kind != EffectSacrifice ||
		effects[1].Kind != EffectChoosePermanent ||
		effects[2].Kind != EffectSearch ||
		!effects[2].SearchSameNameAsChosenObject ||
		effects[3].Kind != EffectPut ||
		effects[4].Kind != EffectShuffle {
		t.Fatalf("effects = %#v", effects)
	}
	for i := range 3 {
		if !effects[i].Exact {
			t.Fatalf("effect %d is not exact", i)
		}
	}
}
