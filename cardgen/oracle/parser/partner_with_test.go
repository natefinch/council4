package parser

import "testing"

// TestParsePartnerWithRecognized confirms the parser recognizes the
// "Partner with <name>" keyword ability, including its pair-fetch reminder, as a
// represented-but-not-simulated partner-with clause and clears the paragraph's
// competing effect, declaration, and condition semantics so downstream stages
// consume only the partner-with identity.
func TestParsePartnerWithRecognized(t *testing.T) {
	t.Parallel()
	source := "Partner with Shabraz, the Skyshark (When this creature enters, target player may put Shabraz into their hand from their library, then shuffle.)"
	document, diagnostics := Parse(source, Context{CardName: "Brallin, Skyshark Rider"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(document.Abilities))
	}
	ability := document.Abilities[0]
	if ability.PartnerWith == nil {
		t.Fatal("PartnerWith clause = nil, want recognized partner-with ability")
	}
	if ability.Sentences != nil || ability.StaticDeclarations != nil ||
		ability.ConditionClauses != nil || ability.EventHistoryConditions != nil {
		t.Fatalf("competing semantics not cleared: %#v", ability)
	}
	assertTextSpan(t, "partner-with name", source, ability.PartnerWith.NameSpan, "Partner with Shabraz, the Skyshark")
}

// TestParsePartnerWithNotTriggeredByPlainPartner confirms the plain "Partner"
// keyword (without a named partner) is not misrecognized as "Partner with", so
// the two-word keyword grammar stays specific to the named-partner form.
func TestParsePartnerWithNotTriggeredByPlainPartner(t *testing.T) {
	t.Parallel()
	keywords := keywordsFor(t, "Partner")
	for _, keyword := range keywords {
		if keyword.Kind == KeywordPartnerWith {
			t.Fatalf("plain Partner misrecognized as Partner with: %+v", keywords)
		}
	}
}
