package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
)

// TestParseSingularSelfCounterStateCondition proves the singular kind-specific
// self-counter condition "this creature has a +1/+1 counter on it" parses to one
// source object-match clause requiring at least one +1/+1 counter, the gate
// Incubation Druid's mana multiplier rider reads. The singular "a <kind> counter"
// presence means one or more counters of that kind.
func TestParseSingularSelfCounterStateCondition(t *testing.T) {
	t.Parallel()
	clause := parseSingleConditionClause(t, "this creature has a +1/+1 counter on it")
	if clause.Predicate != ConditionPredicateObjectMatches {
		t.Fatalf("predicate = %v, want object-matches", clause.Predicate)
	}
	if clause.ObjectBinding != ConditionObjectBindingSource {
		t.Fatalf("object binding = %v, want source", clause.ObjectBinding)
	}
	if !clause.Selection.CounterKindKnown ||
		clause.Selection.CounterKind != counter.PlusOnePlusOne ||
		clause.Selection.CounterCountAtLeast != 1 {
		t.Fatalf("selection = %#v, want +1/+1 count >= 1", clause.Selection)
	}
}
