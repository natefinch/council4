package parser

import "testing"

// TestParseAnyOpponentLostLifeThisTurnCondition covers the reusable
// "an opponent lost N or more life this turn" intervening-if recognizer, both
// for the fully typed predicate on the exact printed wording (Bloodchief
// Ascension, Sygg) and fail-closed behavior on near-miss wording.
func TestParseAnyOpponentLostLifeThisTurnCondition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		condition     string
		wantThreshold int
	}{
		{
			name:          "opponent lost two or more life",
			condition:     "an opponent lost 2 or more life this turn",
			wantThreshold: 2,
		},
		{
			name:          "opponent lost three or more life",
			condition:     "an opponent lost 3 or more life this turn",
			wantThreshold: 3,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			clause := parseSingleConditionClause(t, test.condition)
			if clause.Predicate != ConditionPredicateAnyOpponentLostLifeThisTurnAtLeast {
				t.Fatalf("predicate = %q, want %q", clause.Predicate, ConditionPredicateAnyOpponentLostLifeThisTurnAtLeast)
			}
			if clause.Threshold != test.wantThreshold {
				t.Fatalf("threshold = %d, want %d", clause.Threshold, test.wantThreshold)
			}
		})
	}
}

// TestParseAnyOpponentLostLifeThisTurnConditionFailClosed confirms near-miss
// wording does not resolve to the life-lost predicate, so unrecognized phrasing
// stays typed as unsupported rather than silently matching the wrong meaning.
func TestParseAnyOpponentLostLifeThisTurnConditionFailClosed(t *testing.T) {
	t.Parallel()
	failing := []string{
		// "N or less" is not the printed "N or more" wording.
		"an opponent lost 2 or less life this turn",
		// Missing the per-turn window.
		"an opponent lost 2 or more life",
		// "gained" is the opposite of "lost".
		"an opponent gained 2 or more life this turn",
		// "you" is controller-scoped, not "an opponent".
		"you lost 2 or more life this turn",
	}
	for _, condition := range failing {
		t.Run(condition, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(
				"At the beginning of each end step, if "+condition+", you may draw a card.",
				Context{},
			)
			for _, ability := range document.Abilities {
				for _, clause := range ability.ConditionClauses {
					if clause.Predicate == ConditionPredicateAnyOpponentLostLifeThisTurnAtLeast {
						t.Fatalf("wording %q matched life-lost predicate", condition)
					}
				}
			}
		})
	}
}
