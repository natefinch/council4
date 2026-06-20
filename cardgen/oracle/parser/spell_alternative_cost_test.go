package parser

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
)

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

func TestParseOverloadAlternativeSpellCost(t *testing.T) {
	t.Parallel()
	source := `Destroy target artifact you don't control.
Overload {4}{R} (You may cast this spell for its overload cost. If you do, change "target" in its text to "each.")`
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 2 {
		t.Fatalf("abilities = %d, want 2", len(document.Abilities))
	}
	ability := document.Abilities[1]
	if ability.Kind != AbilitySpellAlternativeCost || ability.AlternativeCost == nil {
		t.Fatalf("ability = %#v, want typed overload cost", ability)
	}
	if ability.AlternativeCost.Kind != SpellAlternativeCostOverload ||
		!ability.AlternativeCost.ReplaceTargetWithEach ||
		!slices.Equal(ability.AlternativeCost.ManaCost, cost.Mana{cost.O(4), cost.R}) {
		t.Fatalf("alternative cost = %#v", ability.AlternativeCost)
	}
	if len(ability.Sentences) != 0 {
		t.Fatalf("overload parsed as resolving content: %#v", ability)
	}
}

func TestParseOverloadFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`Overload {4}{R} (Change "target" to "each.")`,
		`Overload {4}{R} (You may cast this spell for its overload cost. If you do, change "target" in its text to "all.")`,
		`Overload {4}{R}.`,
		`Overload`,
		`Overload {4}{R} (You may cast this spell for its overload cost. If you do, change "target" in its text to "each.") extra`,
	} {
		document, _ := Parse(source, Context{InstantOrSorcery: true})
		if len(document.Abilities) != 1 {
			t.Fatalf("%q abilities = %d, want 1", source, len(document.Abilities))
		}
		if document.Abilities[0].AlternativeCost != nil {
			t.Fatalf("%q unexpectedly recognized as overload", source)
		}
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
