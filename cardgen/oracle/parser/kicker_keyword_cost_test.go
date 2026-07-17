package parser

import "testing"

func TestParseKickerSacrificeCostAsTypedKeyword(t *testing.T) {
	t.Parallel()
	source := "Kicker—Sacrifice a creature. (You may sacrifice a creature in addition to any other costs as you cast this spell.)"
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(document.Abilities))
	}
	ability := document.Abilities[0]
	if ability.Kind != AbilityStatic || ability.AbilityWord != nil {
		t.Fatalf("ability = %#v, want static keyword with no ability word", ability)
	}
	if len(ability.SemanticKeywords) != 1 {
		t.Fatalf("keywords = %#v, want one", ability.SemanticKeywords)
	}
	keyword := ability.SemanticKeywords[0]
	if keyword.Kind != KeywordKicker || keyword.KickerCost == nil {
		t.Fatalf("keyword = %#v, want typed Kicker cost", keyword)
	}
	components := keyword.KickerCost.Components
	if len(components) != 1 ||
		components[0].Kind != CostComponentSacrifice ||
		components[0].ObjectNoun != ObjectNounCreature ||
		!components[0].AmountKnown ||
		components[0].AmountValue != 1 {
		t.Fatalf("components = %#v, want sacrifice one creature", components)
	}
}
