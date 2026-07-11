package parser

import "testing"

func TestParseQuestForRenewalUntap(t *testing.T) {
	t.Parallel()
	const source = "As long as there are four or more quest counters on this enchantment, untap all creatures you control during each other player's untap step."
	document, diagnostics := Parse(source, Context{CardName: "Quest for Renewal"})
	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		t.Fatalf("parse = %#v, diagnostics = %#v", document, diagnostics)
	}
	if !document.Abilities[0].QuestForRenewalUntap {
		t.Fatalf("ability text %q was not recognized", document.Abilities[0].Text)
	}
}
