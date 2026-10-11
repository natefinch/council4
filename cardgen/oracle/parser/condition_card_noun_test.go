package parser

import (
	"slices"
	"testing"
)

func TestObservedCardSubtypeAndAdjectiveConditions(t *testing.T) {
	for _, noun := range []string{"Forest card", "Zombie card", "Goblin permanent card", "snow"} {
		t.Run(noun, func(t *testing.T) {
			document, diagnostics := Parse("Reveal the top card of your library. If it's a "+noun+", draw a card.",
				Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			ability := document.Abilities[0]
			if len(ability.ConditionClauses) != 1 || ability.ConditionClauses[0].Predicate != ConditionPredicateObjectMatches {
				t.Fatalf("typed card noun was not fully modeled: %#v", ability.ConditionClauses)
			}
			selection := ability.ConditionClauses[0].Selection
			if len(ability.SemanticReferences) != 1 || ability.SemanticReferences[0].ProducerClauseID != 1 {
				t.Fatalf("condition did not retain its actual observation: %#v", ability.SemanticReferences)
			}
			if noun == "snow" {
				if len(selection.Supertypes) != 1 {
					t.Fatalf("snow adjective lost its supertype: %#v", selection)
				}
			} else if len(selection.SubtypesAny) != 1 {
				t.Fatalf("subtype noun lost its exact subtype: %#v", selection)
			}
			if noun == "Goblin permanent card" && len(selection.RequiredTypesAny) != 6 {
				t.Fatalf("permanent qualifier was dropped: %#v", selection)
			}
		})
	}
	for _, noun := range []string{"beautiful snow", "Zombie beautiful card", "Goblin beautiful permanent card"} {
		t.Run("unmodeled/"+noun, func(t *testing.T) {
			document, _ := Parse("Reveal the top card of your library. If it's a "+noun+", draw a card.",
				Context{InstantOrSorcery: true})
			for _, condition := range document.Abilities[0].ConditionClauses {
				if condition.Predicate == ConditionPredicateObjectMatches {
					t.Fatal("unmodeled modifier acquired a typed card condition")
				}
			}
		})
	}
}

func TestQualifiedCardConditionCommaIsNounOwned(t *testing.T) {
	for _, noun := range []string{"noncreature, nonland card", "nonland, noncreature card"} {
		t.Run(noun, func(t *testing.T) {
			document, diagnostics := Parse("Reveal the top card of your library. If it's a "+noun+", draw a card.",
				Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			ability := document.Abilities[0]
			if len(ability.ConditionClauses) != 1 || len(ability.ConditionSegments) != 1 ||
				len(ability.SemanticReferences) != 1 || ability.SemanticReferences[0].ProducerClauseID != 1 {
				t.Fatalf("condition lost noun or exact subject ownership: %#v", ability.ConditionClauses)
			}
			selection := ability.ConditionClauses[0].Selection
			if len(selection.RequiredTypesAny) != 0 || len(selection.AnyOf) != 0 ||
				len(selection.ExcludedTypes) != 2 ||
				!slices.Contains(selection.ExcludedTypes, TriggerCardTypeCreature) ||
				!slices.Contains(selection.ExcludedTypes, TriggerCardTypeLand) {
				t.Fatalf("comma exclusions were not conjunctive: %#v", selection)
			}
		})
	}
	for _, noun := range []string{"creature", "noncreature card", "noncreature, beautiful card", "noncreature, nonland card with mystery"} {
		t.Run("boundary/"+noun, func(t *testing.T) {
			document, _ := Parse("Reveal the top card of your library. If it's a "+noun+", draw a card.",
				Context{InstantOrSorcery: true})
			for _, clause := range document.Abilities[0].ConditionClauses {
				if clause.Predicate == ConditionPredicateObjectMatches &&
					len(clause.Selection.ExcludedTypes) == 2 {
					t.Fatal("unmodeled qualifier or ordinary body comma acquired the conjunctive card noun")
				}
			}
		})
	}
}
