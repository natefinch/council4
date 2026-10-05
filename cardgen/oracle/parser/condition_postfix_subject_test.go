package parser

import "testing"

func TestPostfixCounterConditionDoesNotBelongToCopyRider(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Target permanent gains hexproof until end of turn. Put a +1/+1 counter on it if it's a creature. Put a loyalty counter on it if it's a planeswalker.",
		Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics=%+v", diagnostics)
	}
	clauses := document.Abilities[0].ConditionClauses
	if len(clauses) != 2 {
		t.Fatalf("conditions=%+v, want two independent postfix gates", clauses)
	}
	for _, clause := range clauses {
		if clause.Predicate != ConditionPredicateObjectMatches || !clause.HasSubjectSpan || clause.SubjectRefID < 0 {
			t.Fatalf("condition lost its typed subject: %+v", clause)
		}
	}
	if clauses[0].SubjectRefID == clauses[1].SubjectRefID {
		t.Fatal("independent postfix predicates share a subject occurrence")
	}
}
