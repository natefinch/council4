package parser

import (
	"reflect"
	"testing"
)

func TestConditionTypeSelections(t *testing.T) {
	t.Parallel()
	tests := []struct {
		noun string
		want ConditionSelection
	}{
		{"land card", ConditionSelection{RequiredTypes: []TriggerCardType{TriggerCardTypeLand}}},
		{"artifact card", ConditionSelection{RequiredTypes: []TriggerCardType{TriggerCardTypeArtifact}}},
		{"creature cards", ConditionSelection{RequiredTypes: []TriggerCardType{TriggerCardTypeCreature}}},
		{"instant card", ConditionSelection{RequiredTypes: []TriggerCardType{TriggerCardTypeInstant}}},
		{"sorcery card", ConditionSelection{RequiredTypes: []TriggerCardType{TriggerCardTypeSorcery}}},
		{"artifact creature card", ConditionSelection{RequiredTypes: []TriggerCardType{TriggerCardTypeArtifact, TriggerCardTypeCreature}}},
		{"artifact or creature card", ConditionSelection{RequiredTypesAny: []TriggerCardType{TriggerCardTypeArtifact, TriggerCardTypeCreature}}},
		{"artifact card or a creature card", ConditionSelection{RequiredTypesAny: []TriggerCardType{TriggerCardTypeArtifact, TriggerCardTypeCreature}}},
		{"instant and/or sorcery card", ConditionSelection{RequiredTypesAny: []TriggerCardType{TriggerCardTypeInstant, TriggerCardTypeSorcery}}},
		{"artifact or creature or enchantment", ConditionSelection{RequiredTypesAny: []TriggerCardType{TriggerCardTypeArtifact, TriggerCardTypeCreature, TriggerCardTypeEnchantment}}},
		{"noncreature card", ConditionSelection{ExcludedTypes: []TriggerCardType{TriggerCardTypeCreature}}},
		{"nonland card", ConditionSelection{ExcludedTypes: []TriggerCardType{TriggerCardTypeLand}}},
		{"noncreature nonland artifact card", ConditionSelection{RequiredTypes: []TriggerCardType{TriggerCardTypeArtifact}, ExcludedTypes: []TriggerCardType{TriggerCardTypeCreature, TriggerCardTypeLand}}},
		{"artifact creature or enchantment card", ConditionSelection{AnyOf: []ConditionSelection{
			{RequiredTypes: []TriggerCardType{TriggerCardTypeArtifact, TriggerCardTypeCreature}},
			{RequiredTypes: []TriggerCardType{TriggerCardTypeEnchantment}},
		}}},
	}
	for _, test := range tests {
		t.Run(test.noun, func(t *testing.T) {
			t.Parallel()
			clause := parseSingleConditionClause(t, "it was a "+test.noun)
			if clause.Predicate != ConditionPredicateObjectMatches || !reflect.DeepEqual(clause.Selection, test.want) {
				t.Fatalf("clause = %#v, want object match %#v", clause, test.want)
			}
		})
	}
}

func TestConditionTypeSelectionNearMisses(t *testing.T) {
	t.Parallel()
	for _, noun := range []string{
		"card", "permanent card with mystery", "noncreature card from a graveyard",
		"artifact creature creature card", "artifact or mysterious creature card",
		"artifact and creature card", "artifact or creature card with flying and haste",
		"noncreature noncreature card", "artifact or card", "artifact or permanent",
		"legendary artifact or creature card", "artifact or legendary creature card",
		"nonland", "noncreature",
	} {
		t.Run(noun, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse("When this creature enters, if it was a "+noun+", draw a card.", Context{})
			for _, ability := range document.Abilities {
				for _, clause := range ability.ConditionClauses {
					if clause.Predicate == ConditionPredicateObjectMatches {
						t.Fatalf("unexpected supported selection: %#v", clause)
					}
				}
			}
		})
	}
}

func TestConditionTypeSelectionCharacteristicStates(t *testing.T) {
	t.Parallel()
	for _, subject := range []string{"this creature", "equipped creature", "enchanted permanent"} {
		t.Run(subject, func(t *testing.T) {
			t.Parallel()
			clause := parseSingleConditionClause(t, subject+" is a nonland artifact")
			want := ConditionSelection{
				RequiredTypes: []TriggerCardType{TriggerCardTypeArtifact},
				ExcludedTypes: []TriggerCardType{TriggerCardTypeLand},
			}
			if subject == "this creature" {
				want.RequiredTypes = append([]TriggerCardType{TriggerCardTypeCreature}, want.RequiredTypes...)
			}
			if clause.Predicate != ConditionPredicateObjectMatches || !reflect.DeepEqual(clause.Selection, want) {
				t.Fatalf("clause = %#v, want %#v", clause, want)
			}
		})
	}
}
