package parser

import "testing"

func TestParseBorderpostAlternativeCost(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"You may pay {1} and return a basic land you control to its owner's hand rather than pay this spell's mana cost.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	if ability.Kind != AbilitySpellAlternativeCost ||
		ability.AlternativeCost == nil ||
		ability.AlternativeCost.Kind != SpellAlternativeCostBorderpost ||
		ability.CostSyntax == nil ||
		len(ability.CostSyntax.Components) != 1 ||
		ability.CostSyntax.Components[0].Kind != CostComponentReturn {
		t.Fatalf("ability = %#v", ability)
	}
}
