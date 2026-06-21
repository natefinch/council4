package parser

import "testing"

func becomeCopyEffect(t *testing.T, name, text string) EffectSyntax {
	t.Helper()
	doc, _ := Parse(text, Context{CardName: name})
	for a := range doc.Abilities {
		ability := &doc.Abilities[a]
		for s := range ability.Sentences {
			for e := range ability.Sentences[s].Effects {
				effect := ability.Sentences[s].Effects[e]
				if effect.Kind == EffectBecomeCopy {
					return effect
				}
			}
		}
	}
	t.Fatalf("no become-a-copy effect parsed for %q", text)
	return EffectSyntax{}
}

func TestParseBecomeCopyRetainsThisAbility(t *testing.T) {
	effect := becomeCopyEffect(t, "Thespian's Stage",
		"{2}, {T}: This land becomes a copy of target land, except it has this ability.")
	if !effect.BecomeCopyRetainsThisAbility {
		t.Error("expected retains-this-ability rider")
	}
	if effect.BecomeCopyUntilEndOfTurn {
		t.Error("did not expect until-end-of-turn duration")
	}
	if len(effect.BecomeCopyAddKeywords) != 0 {
		t.Errorf("add keywords = %v, want none", effect.BecomeCopyAddKeywords)
	}
}

func TestParseBecomeCopyUntilEndOfTurn(t *testing.T) {
	effect := becomeCopyEffect(t, "Mirage Mirror",
		"{2}: This artifact becomes a copy of target artifact, creature, enchantment, or land until end of turn.")
	if !effect.BecomeCopyUntilEndOfTurn {
		t.Error("expected until-end-of-turn duration")
	}
	if effect.BecomeCopyRetainsThisAbility {
		t.Error("did not expect retains-this-ability rider")
	}
}

func TestParseBecomeCopyKeywordRider(t *testing.T) {
	effect := becomeCopyEffect(t, "Test Copier",
		"{2}, {T}: This creature becomes a copy of target creature, except it has flying.")
	if len(effect.BecomeCopyAddKeywords) != 1 || effect.BecomeCopyAddKeywords[0] != KeywordFlying {
		t.Errorf("add keywords = %v, want [flying]", effect.BecomeCopyAddKeywords)
	}
}
