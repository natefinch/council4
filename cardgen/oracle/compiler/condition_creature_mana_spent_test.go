package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// TestCompileEventSpellCreatureManaSpentCondition confirms the "mana from
// creatures spent to cast" cast-trigger predicate maps from the typed parser
// clause to its compiled counterpart, carrying the numeric threshold. This
// underpins Inga and Esika's "three or more mana from creatures" draw trigger.
func TestCompileEventSpellCreatureManaSpentCondition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		clause        parser.ConditionClause
		wantPredicate ConditionPredicate
		wantThreshold int
	}{
		{
			name: "three or more creature mana",
			clause: parser.ConditionClause{
				Predicate: parser.ConditionPredicateEventSpellCreatureManaSpentToCastAtLeast,
				Threshold: 3,
			},
			wantPredicate: ConditionPredicateEventSpellCreatureManaSpentToCastAtLeast,
			wantThreshold: 3,
		},
		{
			name: "two or more creature mana",
			clause: parser.ConditionClause{
				Predicate: parser.ConditionPredicateEventSpellCreatureManaSpentToCastAtLeast,
				Threshold: 2,
			},
			wantPredicate: ConditionPredicateEventSpellCreatureManaSpentToCastAtLeast,
			wantThreshold: 2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var condition CompiledCondition
			compileConditionClause(&condition, &test.clause)
			if condition.Predicate != test.wantPredicate {
				t.Fatalf("predicate = %v, want %v", condition.Predicate, test.wantPredicate)
			}
			if condition.Threshold != test.wantThreshold {
				t.Fatalf("threshold = %d, want %d", condition.Threshold, test.wantThreshold)
			}
		})
	}
}
