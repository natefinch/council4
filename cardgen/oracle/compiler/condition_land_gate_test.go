package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// TestCompileHideawayGateConditions confirms the three Hideaway play-gate
// predicates map from typed parser clauses to their compiled counterparts,
// carrying the numeric threshold for the two comparison predicates.
func TestCompileHideawayGateConditions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		clause        parser.ConditionClause
		wantPredicate ConditionPredicate
		wantThreshold int
	}{
		{
			name: "opponent damage this turn",
			clause: parser.ConditionClause{
				Predicate: parser.ConditionPredicateAnyOpponentDealtDamageThisTurnAtLeast,
				Threshold: 7,
			},
			wantPredicate: ConditionPredicateAnyOpponentDealtDamageThisTurnAtLeast,
			wantThreshold: 7,
		},
		{
			name: "any library size at most",
			clause: parser.ConditionClause{
				Predicate: parser.ConditionPredicateAnyLibrarySizeAtMost,
				Threshold: 20,
			},
			wantPredicate: ConditionPredicateAnyLibrarySizeAtMost,
			wantThreshold: 20,
		},
		{
			name: "all players hand empty",
			clause: parser.ConditionClause{
				Predicate: parser.ConditionPredicateAllPlayersHandEmpty,
			},
			wantPredicate: ConditionPredicateAllPlayersHandEmpty,
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
