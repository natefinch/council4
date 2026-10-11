package parser

import "testing"

func TestObservedQualifiedUnionContractionIdentity(t *testing.T) {
	for _, noun := range []string{
		"permanent card with mana value 3 or less",
		"artifact or creature card with mana value 3 or less",
		"noncreature, nonland card",
	} {
		t.Run(noun, func(t *testing.T) {
			document, diagnostics := Parse("When this creature dies, reveal the top card of your library. You may put that card onto the battlefield if it's a "+noun+". Otherwise, put that card into your hand.", Context{})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			ability := document.Abilities[0]
			if len(ability.ConditionClauses) != 1 {
				t.Fatalf("conditions=%#v", ability.ConditionClauses)
			}
			condition := ability.ConditionClauses[0]
			if !condition.HasSubjectSpan || condition.SubjectRefID < 0 {
				t.Fatalf("qualified union lost condition subject identity: %#v", condition)
			}
			found := false
			for _, reference := range ability.SemanticReferences {
				if reference.NodeID == condition.SubjectRefID {
					found = reference.ProducerClauseID == ability.Sentences[0].Effects[0].ClauseID
				}
			}
			if !found {
				t.Fatal("condition does not own the exact observation producer")
			}
		})
	}
}
