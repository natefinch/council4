package parser

import (
	"slices"
	"testing"
)

func TestParseRavenousKeywordWithReminderText(t *testing.T) {
	t.Parallel()
	source := "Ravenous (This creature enters with X +1/+1 counters on it. If X is 5 or more, draw a card when it enters.)"
	keywords := keywordsFor(t, source)
	if len(keywords) != 1 {
		t.Fatalf("keywords = %+v; want one", keywords)
	}
	if got := keywords[0]; got.Kind != KeywordRavenous ||
		got.Parameter.Kind != KeywordParameterNone ||
		got.Text != "Ravenous" {
		t.Fatalf("keyword = %+v; want Ravenous with reminder text excluded from keyword span", got)
	}
}

func TestParseDamagedPlayerControlledArtifactOrEnchantmentTarget(t *testing.T) {
	t.Parallel()
	target := singleTarget(t, "Destroy target artifact or enchantment that player controls.")
	if !target.Exact ||
		target.Selection.Controller != SelectionControllerThatPlayer ||
		!slices.Equal(target.Selection.RequiredTypesAny, []CardType{CardTypeArtifact, CardTypeEnchantment}) {
		t.Fatalf("target = %#v; want exact artifact-or-enchantment target controlled by that player", target)
	}
}
