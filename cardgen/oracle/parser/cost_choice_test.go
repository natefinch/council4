package parser

import "testing"

// spellAdditionalCostComponents parses an instant whose only ability is an "As
// an additional cost to cast this spell, ..." clause and returns its typed cost
// components.
func spellAdditionalCostComponents(t *testing.T, costText string) []CostComponent {
	t.Helper()
	source := "As an additional cost to cast this spell, " + costText + ".\nDraw a card."
	document, _ := Parse(source, Context{InstantOrSorcery: true})
	for ai := range document.Abilities {
		ability := &document.Abilities[ai]
		if ability.Kind == AbilitySpellAdditionalCost && ability.CostSyntax != nil {
			return ability.CostSyntax.Components
		}
	}
	t.Fatalf("no spell additional-cost ability parsed from %q", source)
	return nil
}

func TestParseAdditionalCostChoice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		cost  string
		kinds []CostComponentKind
	}{
		{
			name:  "sacrifice or discard",
			cost:  "sacrifice an artifact or discard a card",
			kinds: []CostComponentKind{CostComponentSacrifice, CostComponentDiscard},
		},
		{
			name:  "discard or pay life",
			cost:  "discard a card or pay 3 life",
			kinds: []CostComponentKind{CostComponentDiscard, CostComponentPayLife},
		},
		{
			name:  "oxford three-way",
			cost:  "sacrifice a creature, discard a card, or pay 4 life",
			kinds: []CostComponentKind{CostComponentSacrifice, CostComponentDiscard, CostComponentPayLife},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			components := spellAdditionalCostComponents(t, test.cost)
			if len(components) != len(test.kinds) {
				t.Fatalf("components = %d, want %d (%#v)", len(components), len(test.kinds), components)
			}
			for i, component := range components {
				if component.Kind != test.kinds[i] {
					t.Fatalf("component %d kind = %v, want %v", i, component.Kind, test.kinds[i])
				}
				if component.ChoiceGroup != 1 {
					t.Fatalf("component %d choice group = %d, want 1", i, component.ChoiceGroup)
				}
			}
		})
	}
}

// TestParseAdditionalCostTypeUnionNotChoice guards that a two-permanent-type
// union keeps its single-component shape and is not split into a choice.
func TestParseAdditionalCostTypeUnionNotChoice(t *testing.T) {
	t.Parallel()
	components := spellAdditionalCostComponents(t, "sacrifice an artifact or creature")
	if len(components) != 1 {
		t.Fatalf("components = %d, want 1 (%#v)", len(components), components)
	}
	if components[0].Kind != CostComponentSacrifice {
		t.Fatalf("kind = %v, want sacrifice", components[0].Kind)
	}
	if components[0].ChoiceGroup != 0 {
		t.Fatalf("choice group = %d, want 0", components[0].ChoiceGroup)
	}
}
