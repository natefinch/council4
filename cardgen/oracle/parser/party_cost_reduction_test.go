package parser

import "testing"

func TestParsePartyCostReduction(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"This spell costs {1} less to cast for each creature in your party.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effect := document.Abilities[0].Sentences[0].Effects[0]
	if effect.Amount.DynamicKind != EffectDynamicAmountPartySize ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormForEach ||
		effect.Amount.Multiplier != 1 ||
		!effect.SourceSpellCostReductionDynamic {
		t.Fatalf("effect = %#v", effect)
	}
}
