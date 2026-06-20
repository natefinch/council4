package parser

import "testing"

func TestParseCommanderControlledAlternativeSpellCost(t *testing.T) {
	t.Parallel()
	source := "If you control a commander, you may cast this spell without paying its mana cost.\nCounter target noncreature spell."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 2 {
		t.Fatalf("abilities = %d, want 2", len(document.Abilities))
	}
	ability := document.Abilities[0]
	if ability.Kind != AbilitySpellAlternativeCost || ability.AlternativeCost == nil {
		t.Fatalf("ability = %#v, want typed alternative spell cost", ability)
	}
	if ability.AlternativeCost.Condition != SpellAlternativeCostConditionControlsCommander ||
		!ability.AlternativeCost.WithoutPayingManaCost {
		t.Fatalf("alternative cost = %#v", ability.AlternativeCost)
	}
	if len(ability.Sentences) != 0 || ability.Optional {
		t.Fatalf("alternative cost parsed as resolving content: %#v", ability)
	}
}

func TestParseCommanderControlledAlternativeSpellCostFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"If you own a commander, you may cast this spell without paying its mana cost.",
		"If you control a commander, cast this spell without paying its mana cost.",
		"If you control your commander, you may cast this spell without paying its mana cost.",
		"If you control a commander, you may cast this spell without paying its mana cost from exile.",
		"If you control a commander, you may cast this spell without paying its mana cost",
		"If you control a commander. You may cast this spell without paying its mana cost.",
		"If you control a commander; you may cast this spell without paying its mana cost.",
		"If you control a commander you may cast this spell without paying its mana cost.",
		"If you control a commander,, you may cast this spell without paying its mana cost.",
		"If you control a commander, you may cast this spell without paying its mana cost..",
		"If you control a commander, you may cast this spell; without paying its mana cost.",
	} {
		document, _ := Parse(source, Context{InstantOrSorcery: true})
		if len(document.Abilities) != 1 {
			t.Fatalf("%q abilities = %d, want 1", source, len(document.Abilities))
		}
		if document.Abilities[0].AlternativeCost != nil {
			t.Fatalf("%q unexpectedly recognized as alternative cost", source)
		}
	}
}
