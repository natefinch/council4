package parser

import "testing"

// TestParseHideawayGateConditions covers the three Hideaway play-gate condition
// recognizers, asserting both the fully typed predicate for the exact printed
// wording and fail-closed behavior on near-miss wording.
func TestParseHideawayGateConditions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		condition     string
		wantPredicate ConditionPredicateKind
		wantThreshold int
	}{
		{
			name:          "opponent dealt seven or more damage",
			condition:     "an opponent was dealt 7 or more damage this turn",
			wantPredicate: ConditionPredicateAnyOpponentDealtDamageThisTurnAtLeast,
			wantThreshold: 7,
		},
		{
			name:          "opponent dealt three or more damage",
			condition:     "an opponent was dealt 3 or more damage this turn",
			wantPredicate: ConditionPredicateAnyOpponentDealtDamageThisTurnAtLeast,
			wantThreshold: 3,
		},
		{
			name:          "a library twenty or fewer",
			condition:     "a library has 20 or fewer cards in it",
			wantPredicate: ConditionPredicateAnyLibrarySizeAtMost,
			wantThreshold: 20,
		},
		{
			name:          "each player empty hand",
			condition:     "each player has no cards in hand",
			wantPredicate: ConditionPredicateAllPlayersHandEmpty,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			clause := parseSingleConditionClause(t, test.condition)
			if clause.Predicate != test.wantPredicate {
				t.Fatalf("predicate = %q, want %q", clause.Predicate, test.wantPredicate)
			}
			if clause.Threshold != test.wantThreshold {
				t.Fatalf("threshold = %d, want %d", clause.Threshold, test.wantThreshold)
			}
		})
	}
}

// TestParseHideawayGateConditionsFailClosed confirms near-miss wording does not
// resolve to a Hideaway gate predicate, so unrecognized phrasing stays typed as
// unsupported rather than silently matching the wrong meaning.
func TestParseHideawayGateConditionsFailClosed(t *testing.T) {
	t.Parallel()
	failing := []string{
		// "N or less" is not the printed "N or more" damage wording.
		"an opponent was dealt 7 or less damage this turn",
		// Missing the per-turn window.
		"an opponent was dealt 7 or more damage",
		// "N or more" is not the printed "N or fewer" library wording.
		"a library has 20 or more cards in it",
		// "your library" is a controller-scoped predicate, not "a library".
		"your library has 20 or fewer cards in it",
		// "a player" is not the universal "each player".
		"a player has no cards in hand",
		// "one card" is not "no cards".
		"each player has one card in hand",
	}
	for _, condition := range failing {
		t.Run(condition, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(
				"When this creature enters, if "+condition+", draw a card.",
				Context{},
			)
			for _, ability := range document.Abilities {
				for _, clause := range ability.ConditionClauses {
					switch clause.Predicate {
					case ConditionPredicateAnyOpponentDealtDamageThisTurnAtLeast,
						ConditionPredicateAnyLibrarySizeAtMost,
						ConditionPredicateAllPlayersHandEmpty:
						t.Fatalf("wording %q matched Hideaway gate predicate %q", condition, clause.Predicate)
					default:
						// Any other predicate (including unsupported) is acceptable:
						// the near-miss wording must not resolve to a gate predicate.
					}
				}
			}
		})
	}
}
