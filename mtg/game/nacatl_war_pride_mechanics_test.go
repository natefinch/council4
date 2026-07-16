package game

import "testing"

func TestMustBeBlockedByExactlyOneStaticBodyComposesBounds(t *testing.T) {
	t.Parallel()
	body := MustBeBlockedByExactlyOneStaticBody
	if body.Text != "This creature must be blocked by exactly one creature if able." {
		t.Fatalf("text = %q", body.Text)
	}
	if len(body.RuleEffects) != 2 ||
		body.RuleEffects[0].Kind != RuleEffectMustBeBlocked ||
		body.RuleEffects[1].Kind != RuleEffectCantBeBlockedByMoreThanOne ||
		!body.RuleEffects[0].AffectedSource ||
		!body.RuleEffects[1].AffectedSource {
		t.Fatalf("rule effects = %#v", body.RuleEffects)
	}
}
