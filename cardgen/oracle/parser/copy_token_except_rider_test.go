package parser

import "testing"

// TestParseCopyTokenExceptKeywordRider verifies the inline copy-token rider
// "except <the token> has <keyword> and it isn't legendary" (Irenicus's Vile
// Duplication) folds into a single copy create that records the granted keyword
// and the drop-legendary flag rather than stranding a separate keyword-grant
// effect.
func TestParseCopyTokenExceptKeywordRider(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Create a token that's a copy of target creature you control, except the token has flying and it isn't legendary.",
		Context{InstantOrSorcery: true, CardName: "Irenicus's Vile Duplication"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 {
		t.Fatalf("effects = %#v, want a single folded copy create", effects)
	}
	effect := effects[0]
	if !effect.TokenCopyOfTarget {
		t.Fatalf("TokenCopyOfTarget = false, want true: %#v", effect)
	}
	if !effect.TokenCopyDropLegendary {
		t.Error("TokenCopyDropLegendary = false, want true")
	}
	if len(effect.TokenCopyGrantKeywords) != 1 || effect.TokenCopyGrantKeywords[0] != KeywordFlying {
		t.Errorf("TokenCopyGrantKeywords = %v, want [flying]", effect.TokenCopyGrantKeywords)
	}
}

// TestParseCopyTokenExceptQuotedAbilityFailsClosed verifies a copy-token rider
// carrying a quoted granted ability ("except it has haste and \"...sacrifice
// this token.\"" — Electroduplicate, Heat Shimmer) fails closed: the quoted
// ability cannot be represented, so the copy is not recognized as exact rather
// than silently dropping the ability and keeping only the keyword.
func TestParseCopyTokenExceptQuotedAbilityFailsClosed(t *testing.T) {
	t.Parallel()
	document, _ := Parse(
		"Create a token that's a copy of target creature you control, except it has haste and \"At the beginning of the end step, sacrifice this token.\"",
		Context{InstantOrSorcery: true, CardName: "Electroduplicate"})
	for _, sentence := range document.Abilities[0].Sentences {
		for _, effect := range sentence.Effects {
			if effect.TokenCopyOfTarget {
				t.Fatalf("quoted-ability rider must fail closed, got recognized copy create with keywords %v", effect.TokenCopyGrantKeywords)
			}
		}
	}
}
