package game

import "testing"

func TestEntersAsCopyExceptionWrappersCompose(t *testing.T) {
	replacement := EntersAsCopyWithOtherAbilities(
		EntersAsCopyWithRetainedName(
			EntersAsCopyReplacement("copy", &Selection{}, true, false, nil, false, nil, nil),
		),
	)
	if !replacement.Replacement.EntersAsCopyRetainName {
		t.Fatal("retained-name copy exception was not set")
	}
	if !replacement.Replacement.EntersAsCopyAddOtherAbilities {
		t.Fatal("other-abilities copy exception was not set")
	}
}

func TestLegendRuleDoesNotApplyStaticBody(t *testing.T) {
	if len(LegendRuleDoesNotApplyStaticBody.RuleEffects) != 1 {
		t.Fatalf("rule effects = %#v, want one", LegendRuleDoesNotApplyStaticBody.RuleEffects)
	}
	effect := LegendRuleDoesNotApplyStaticBody.RuleEffects[0]
	if effect.Kind != RuleEffectLegendRuleDoesNotApply || effect.AffectedPlayer != PlayerYou {
		t.Fatalf("rule effect = %#v, want controller legend-rule exemption", effect)
	}
}
