package parser

import (
	"slices"
	"testing"
)

// TestParseConditionControlsGreatestManaValueInGroup recognizes the
// intervening-if wording "you control the <noun> with the greatest mana value or
// tied for the greatest mana value" and captures the filtered group as a typed
// Selection. The predicate is generic over the noun so it is not hardcoded to
// artifacts: any parseable group filter is carried through to the runtime
// aggregate comparison.
func TestParseConditionControlsGreatestManaValueInGroup(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		condition string
		cardType  TriggerCardType
	}{
		{
			name:      "artifact group (Padeem)",
			condition: "you control the artifact with the greatest mana value or tied for the greatest mana value",
			cardType:  TriggerCardTypeArtifact,
		},
		{
			name:      "creature group is equally recognized",
			condition: "you control the creature with the greatest mana value or tied for the greatest mana value",
			cardType:  TriggerCardTypeCreature,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			clause := parseSingleConditionClause(t, test.condition)
			if clause.Predicate != ConditionPredicateControlsGreatestManaValueInGroup {
				t.Fatalf("clause = %#v, want greatest-mana-value predicate", clause)
			}
			if !slices.Equal(clause.Selection.RequiredTypes, []TriggerCardType{test.cardType}) {
				t.Fatalf("selection = %#v, want card type %s", clause.Selection, test.cardType)
			}
		})
	}
}

// TestParseConditionControlsGreatestManaValueFailsClosed rejects near-miss
// wordings so the aggregate mana-value comparison never fires on text it does
// not exactly describe. The parser owns the exact wording; anything else fails
// closed rather than approximating a different measurement (power vs. mana
// value), a different superlative ("highest"), or a one-sided comparison that
// drops the tie clause.
func TestParseConditionControlsGreatestManaValueFailsClosed(t *testing.T) {
	t.Parallel()
	conditions := []string{
		"you control the artifact with the greatest mana value",
		"you control the artifact with the greatest power or tied for the greatest power",
		"you control the artifact with the highest mana value or tied for the highest mana value",
		"you control an artifact with the greatest mana value or tied for the greatest mana value",
		"you control the artifact with the greatest mana value or tied for the greatest power",
	}
	for _, condition := range conditions {
		t.Run(condition, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(
				"When this creature enters, if "+condition+", draw a card.",
				Context{},
			)
			if len(document.Abilities) != 1 {
				return
			}
			for _, clause := range document.Abilities[0].ConditionClauses {
				if clause.Predicate == ConditionPredicateControlsGreatestManaValueInGroup {
					t.Fatalf("condition %q produced greatest-mana-value clause %#v, want none", condition, clause)
				}
			}
		})
	}
}
