package parser

import "testing"

// TestParseCommanderCastCountAnthem verifies that the group anthem "Creatures
// you control get +1/+1 for each time you've cast your commander from the
// command zone this game." types to the commander-cast count, for-each dynamic
// form, backing the command-zone-cast anthem family (Commander's Insignia;
// Vanguard of the Restless).
func TestParseCommanderCastCountAnthem(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Creatures you control get +1/+1 for each time you've cast your commander from the command zone this game.",
		Context{CardName: "Commander's Insignia"},
	)
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
	if effect.Kind != EffectModifyPT ||
		effect.Amount.DynamicKind != EffectDynamicAmountCommanderCastCount ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormForEach {
		t.Fatalf("amount = %#v kind = %v", effect.Amount, effect.Kind)
	}
}
