package cardgen

import "testing"

func TestCantBeBlockedRestrictionTimingFailsClosed(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Target creature can't be blocked this turn if it's tapped.",
		"Draw a card. Target creature can't be blocked this turn if it's tapped.",
		"Choose one —\n• Target creature can't be blocked this turn if it's tapped.\n• Draw a card.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			assertCardUnsupported(t, &ScryfallCard{Name: "Restriction Timing", Layout: "normal", TypeLine: "Instant", OracleText: text},
				"trailing can't-be-blocked restriction predicate")
		})
	}
	for _, text := range []string{
		"Target creature can't be blocked this turn.",
		"If you control a creature, target creature can't be blocked this turn.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			assertCardPaths(t, &ScryfallCard{Name: "Resolving Grant", Layout: "normal", TypeLine: "Instant", OracleText: text},
				"Primitive.(game.ApplyRule)", "Kind = game.RuleEffectCantBeBlocked")
		})
	}
	assertCardPaths(t, &ScryfallCard{Name: "Other Postfix", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "Draw a card if you control a creature."},
		"Primitive.(game.Draw)", "Condition.Val.Condition.Val.ControlsMatching")
}
