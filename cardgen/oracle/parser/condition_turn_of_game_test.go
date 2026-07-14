package parser

import "testing"

// TestParseControllerTurnOfGameCondition covers the per-player turn-ordinal
// wording ("it's your first, second, or third turn of the game"). The contiguous
// ordinal run collapses to an at-most threshold equal to the highest ordinal,
// and the comma-separated run survives condition-clause splitting because the
// clause-end override keeps the whole run in one clause.
func TestParseControllerTurnOfGameCondition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		condition string
		threshold int
	}{
		{"first only", "it's your first turn of the game", 1},
		{"first or second", "it's your first or second turn of the game", 2},
		{"first second or third", "it's your first, second, or third turn of the game", 3},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			clause := parseSingleConditionClause(t, test.condition)
			if clause.Predicate != ConditionPredicateControllerTurnOfGameAtMost {
				t.Fatalf("predicate = %s, want controller-turn-of-game", clause.Predicate)
			}
			if clause.Threshold != test.threshold {
				t.Fatalf("threshold = %d, want %d", clause.Threshold, test.threshold)
			}
		})
	}
}

// TestParseControllerTurnOfGameConditionFailsClosed confirms wordings that are
// not a contiguous ordinal run anchored at "first" are not recognized as the
// turn-of-game predicate, so they fall through rather than compiling to a wrong
// threshold.
func TestParseControllerTurnOfGameConditionFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		condition string
	}{
		{"not anchored at first", "it's your second turn of the game"},
		{"gap in run", "it's your first or third turn of the game"},
		{"missing tail", "it's your first turn"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(
				"When this creature enters, if "+test.condition+", draw a card.",
				Context{},
			)
			if len(document.Abilities) == 0 {
				return
			}
			for _, clause := range document.Abilities[0].ConditionClauses {
				if clause.Predicate == ConditionPredicateControllerTurnOfGameAtMost {
					t.Fatalf("condition %q was recognized as turn-of-game predicate", test.condition)
				}
			}
		})
	}
}
