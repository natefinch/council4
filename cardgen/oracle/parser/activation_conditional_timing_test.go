package parser

import "testing"

// TestConditionalActivationTimingTailSplit proves that an "Activate only if
// <condition> and only <timing>" sentence is split: the trailing "and only
// <timing>" yields a typed activation-timing restriction while the leading
// "only if <condition>" prefix is recognized as the activation condition. The
// conjoined timing tail must not fracture or block the condition.
func TestConditionalActivationTimingTailSplit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		source     string
		wantTiming ActivationRestrictionKind
	}{
		{
			name:       "and only as a sorcery",
			source:     "{2}{W}: Draw a card. Activate only if you control an enchantment and only as a sorcery.",
			wantTiming: ActivationRestrictionSorceryTiming,
		},
		{
			name:       "and only once each turn",
			source:     "{B}: Draw a card. Activate only if an opponent has three or more poison counters and only once each turn.",
			wantTiming: ActivationRestrictionFrequency,
		},
		{
			name:       "and only during your turn",
			source:     "{2}{W}: Draw a card. Activate only if you control no creatures and only during your turn.",
			wantTiming: ActivationRestrictionPlayerTurn,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ability := parseSingleAbility(t, tc.source, Context{})
			if len(ability.ActivationRestrictions) != 1 {
				t.Fatalf("restrictions = %#v, want one timing tail", ability.ActivationRestrictions)
			}
			if got := ability.ActivationRestrictions[0].Kind; got != tc.wantTiming {
				t.Fatalf("restriction kind = %q, want %q", got, tc.wantTiming)
			}
			var onlyIf int
			for _, boundary := range ability.ConditionBoundaries {
				if boundary.Kind == ConditionIntroOnlyIf {
					onlyIf++
				}
			}
			if onlyIf != 1 {
				t.Fatalf("only-if boundaries = %d, want one", onlyIf)
			}
			if len(ability.ConditionClauses) != 1 {
				t.Fatalf("condition clauses = %#v, want the recognized gate", ability.ConditionClauses)
			}
		})
	}
}

// TestBareConditionalActivationStaysCondition proves a bare "Activate only if
// <condition>" with no conjoined timing tail emits no activation-timing
// restriction, so the pure condition path is unchanged.
func TestBareConditionalActivationStaysCondition(t *testing.T) {
	t.Parallel()
	ability := parseSingleAbility(t,
		"{T}: Draw a card. Activate only if you have 10 or more life.", Context{})
	if len(ability.ActivationRestrictions) != 0 {
		t.Fatalf("restrictions = %#v, want none for a bare condition", ability.ActivationRestrictions)
	}
	if len(ability.ConditionClauses) != 1 {
		t.Fatalf("condition clauses = %#v, want the recognized gate", ability.ConditionClauses)
	}
}
