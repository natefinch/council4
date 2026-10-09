package parser

import "testing"

func TestReferencedBattlefieldDestinationExactControl(t *testing.T) {
	for _, test := range []struct {
		clause                   string
		exact, controller, owner bool
	}{
		{"Put that card onto the battlefield under your control.", true, true, false},
		{"Put that card onto the battlefield.", true, false, false},
		{"Return it to the battlefield under your control.", true, true, false},
		{"Return it to the battlefield under its owner's control.", true, false, true},
		{"Return it to the battlefield converted under its owner's control.", true, false, true},
		{"Return it to the battlefield transformed under your control.", true, true, false},
		{"Return it to the battlefield tapped and transformed under its owner's control.", true, false, true},
		{"Return the chosen card to the battlefield tapped.", false, false, false},
		{"Put that card onto the battlefield under that player's control.", false, false, false},
		{"Put that card onto the battlefield under target opponent's control.", false, false, false},
		{"Put that card onto the battlefield under your control and an opponent's control.", false, true, false},
		{"Return it to the battlefield under your control or its owner's control.", false, true, true},
	} {
		t.Run(test.clause, func(t *testing.T) {
			document, diagnostics := Parse(
				"Whenever an opponent sacrifices a nontoken permanent, "+test.clause, Context{CardName: "Recipient"})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			ability := document.Abilities[0]
			if len(ability.Sentences) != 1 || len(ability.Sentences[0].Effects) != 1 {
				t.Fatalf("unexpected effects: %+v", ability.Sentences)
			}
			effect := ability.Sentences[0].Effects[0]
			if effect.Exact != test.exact || effect.UnderYourControl != test.controller ||
				effect.UnderOwnersControl != test.owner {
				t.Fatalf("exact/controller/owner=%v/%v/%v want=%v/%v/%v", effect.Exact,
					effect.UnderYourControl, effect.UnderOwnersControl, test.exact, test.controller, test.owner)
			}
		})
	}
}
