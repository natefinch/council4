package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// TestCompileControlComparisonConditionFromTypedNodes drives the cross-player
// control-count comparison compiler with constructed typed parser nodes only.
// The compiler derives the comparison meaning from the typed ControlComparison
// scopes and Selection, never from Oracle source text.
func TestCompileControlComparisonConditionFromTypedNodes(t *testing.T) {
	t.Parallel()
	clause := parser.ConditionClause{
		Predicate: parser.ConditionPredicateControlComparison,
		Selection: parser.ConditionSelection{
			RequiredTypes: []parser.TriggerCardType{parser.TriggerCardTypeLand},
		},
		ControlComparison: parser.ConditionControlComparison{
			LeftScope:  parser.ConditionControlScopeAnyOpponent,
			RightScope: parser.ConditionControlScopeController,
			Greater:    true,
		},
	}
	var condition CompiledCondition
	compileConditionClause(&condition, &clause)
	if condition.Predicate != ConditionPredicateControlComparison {
		t.Fatalf("predicate = %v, want control comparison", condition.Predicate)
	}
	if condition.ControlComparisonLeft != ConditionComparisonScopeAnyOpponent ||
		condition.ControlComparisonRight != ConditionComparisonScopeController ||
		!condition.ControlComparisonGreater {
		t.Fatalf("condition = %#v, want opponent>controller greater", condition)
	}
}

// TestCompileControlComparisonConditionFailsClosed rejects comparisons whose two
// sides do not contrast the controller against an opponent scope, leaving the
// predicate unsupported.
func TestCompileControlComparisonConditionFailsClosed(t *testing.T) {
	t.Parallel()
	clause := parser.ConditionClause{
		Predicate: parser.ConditionPredicateControlComparison,
		Selection: parser.ConditionSelection{
			RequiredTypes: []parser.TriggerCardType{parser.TriggerCardTypeLand},
		},
		ControlComparison: parser.ConditionControlComparison{
			LeftScope:  parser.ConditionControlScopeAnyOpponent,
			RightScope: parser.ConditionControlScopeEachOpponent,
			Greater:    true,
		},
	}
	var condition CompiledCondition
	compileConditionClause(&condition, &clause)
	if condition.Predicate == ConditionPredicateControlComparison {
		t.Fatalf("condition = %#v, want unsupported predicate", condition)
	}
}
