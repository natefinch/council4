package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

func entersAsCopyEffect(t *testing.T, name, text string) EffectSyntax {
	t.Helper()
	doc, _ := Parse(text, Context{CardName: name})
	for a := range doc.Abilities {
		ability := &doc.Abilities[a]
		if ability.Kind != AbilityReplacement {
			continue
		}
		for s := range ability.Sentences {
			for e := range ability.Sentences[s].Effects {
				effect := ability.Sentences[s].Effects[e]
				if effect.EntersAsCopy {
					return effect
				}
			}
		}
	}
	t.Fatalf("no enters-as-copy effect parsed for %q", text)
	return EffectSyntax{}
}

func TestParseEntersAsCopyClone(t *testing.T) {
	effect := entersAsCopyEffect(t, "Clone",
		"You may have Clone enter the battlefield as a copy of any creature on the battlefield.")
	if !effect.EntersAsCopyOptional {
		t.Error("expected optional copy")
	}
	if effect.EntersAsCopyNotLegendary {
		t.Error("did not expect not-legendary rider")
	}
	if effect.Selection.Controller != SelectionControllerAny {
		t.Errorf("controller = %v, want any", effect.Selection.Controller)
	}
}

func TestParseEntersAsCopyControlledNotLegendary(t *testing.T) {
	effect := entersAsCopyEffect(t, "Spark Double",
		"You may have Spark Double enter the battlefield as a copy of a creature you control, except it isn't legendary.")
	if !effect.EntersAsCopyNotLegendary {
		t.Error("expected not-legendary rider")
	}
	if effect.Selection.Controller != SelectionControllerYou {
		t.Errorf("controller = %v, want you", effect.Selection.Controller)
	}
}

func TestParseEntersAsCopyAddArtifact(t *testing.T) {
	effect := entersAsCopyEffect(t, "Phyrexian Metamorph",
		"You may have Phyrexian Metamorph enter the battlefield as a copy of any artifact or creature on the battlefield, except it's an artifact in addition to its other types.")
	if len(effect.EntersAsCopyAddTypes) != 1 || effect.EntersAsCopyAddTypes[0] != types.Artifact {
		t.Errorf("add types = %v, want [artifact]", effect.EntersAsCopyAddTypes)
	}
}

func TestParseEntersAsCopyUntilEndOfTurnKeywordRider(t *testing.T) {
	effect := entersAsCopyEffect(t, "Cursed Mirror",
		"As this artifact enters, you may have it become a copy of any creature on the battlefield until end of turn, except it has haste.")
	if !effect.EntersAsCopyOptional {
		t.Error("expected optional copy")
	}
	if !effect.EntersAsCopyUntilEndOfTurn {
		t.Error("expected until-end-of-turn copy duration")
	}
	if len(effect.EntersAsCopyAddKeywords) != 1 || effect.EntersAsCopyAddKeywords[0] != KeywordHaste {
		t.Errorf("add keywords = %v, want [haste]", effect.EntersAsCopyAddKeywords)
	}
	if len(effect.Selection.RequiredTypesAny) != 1 || effect.Selection.RequiredTypesAny[0] != CardTypeCreature {
		t.Errorf("required types = %v, want [creature]", effect.Selection.RequiredTypesAny)
	}
}

func TestParseEntersAsCopyConditionalCounters(t *testing.T) {
	effect := entersAsCopyEffect(t, "Spark Double",
		"You may have this creature enter as a copy of a creature or planeswalker you control, except it enters with an additional +1/+1 counter on it if it's a creature, it enters with an additional loyalty counter on it if it's a planeswalker, and it isn't legendary.")
	if !effect.EntersAsCopyOptional {
		t.Error("expected optional copy")
	}
	if !effect.EntersAsCopyNotLegendary {
		t.Error("expected not-legendary rider")
	}
	if effect.Selection.Controller != SelectionControllerYou {
		t.Errorf("controller = %v, want you", effect.Selection.Controller)
	}
	if len(effect.Selection.RequiredTypesAny) != 2 {
		t.Errorf("required-any types = %v, want creature/planeswalker", effect.Selection.RequiredTypesAny)
	}
	got := effect.EntersAsCopyConditionalCounters
	if len(got) != 2 {
		t.Fatalf("conditional counters = %+v, want two", got)
	}
	if got[0].Kind != counter.PlusOnePlusOne || got[0].Amount != 1 || got[0].IfType != types.Creature {
		t.Errorf("counter[0] = %+v, want +1/+1 if creature", got[0])
	}
	if got[1].Kind != counter.Loyalty || got[1].Amount != 1 || got[1].IfType != types.Planeswalker {
		t.Errorf("counter[1] = %+v, want loyalty if planeswalker", got[1])
	}
}

func TestParseEntersAsCopyAddSubtype(t *testing.T) {
	effect := entersAsCopyEffect(t, "Mockingbird",
		"You may have this creature enter as a copy of any creature on the battlefield, except it's a Bird in addition to its other types and it has flying.")
	if len(effect.EntersAsCopyAddSubtypes) != 1 || effect.EntersAsCopyAddSubtypes[0] != types.Bird {
		t.Errorf("add subtypes = %v, want [Bird]", effect.EntersAsCopyAddSubtypes)
	}
	if len(effect.EntersAsCopyAddKeywords) != 1 || effect.EntersAsCopyAddKeywords[0] != KeywordFlying {
		t.Errorf("add keywords = %v, want [flying]", effect.EntersAsCopyAddKeywords)
	}
	if len(effect.EntersAsCopyAddTypes) != 0 {
		t.Errorf("add types = %v, want none", effect.EntersAsCopyAddTypes)
	}
}

func TestParseEntersAsCopyAddSubtypeAndCardTypes(t *testing.T) {
	effect := entersAsCopyEffect(t, "Synth Infiltrator",
		"You may have this creature enter as a copy of any creature on the battlefield, except it's a Synth artifact creature in addition to its other types.")
	if len(effect.EntersAsCopyAddSubtypes) != 1 || effect.EntersAsCopyAddSubtypes[0] != types.Synth {
		t.Errorf("add subtypes = %v, want [Synth]", effect.EntersAsCopyAddSubtypes)
	}
	wantTypes := []types.Card{types.Artifact, types.Creature}
	if len(effect.EntersAsCopyAddTypes) != 2 ||
		effect.EntersAsCopyAddTypes[0] != wantTypes[0] || effect.EntersAsCopyAddTypes[1] != wantTypes[1] {
		t.Errorf("add types = %v, want %v", effect.EntersAsCopyAddTypes, wantTypes)
	}
}

func hasEntersAsCopyEffect(t *testing.T, name, text string) bool {
	t.Helper()
	doc, _ := Parse(text, Context{CardName: name})
	for a := range doc.Abilities {
		ability := &doc.Abilities[a]
		for s := range ability.Sentences {
			for e := range ability.Sentences[s].Effects {
				if ability.Sentences[s].Effects[e].EntersAsCopy {
					return true
				}
			}
		}
	}
	return false
}

func TestParseEntersAsCopyFailsClosed(t *testing.T) {
	cases := []struct{ name, text string }{
		{"Essence of the Wild", "Creatures you control enter as a copy of this creature."},
		{"Body Double", "You may have this creature enter as a copy of any creature card in a graveyard."},
	}
	for _, tc := range cases {
		if hasEntersAsCopyEffect(t, tc.name, tc.text) {
			t.Errorf("%s: expected enters-as-copy to fail closed", tc.name)
		}
	}
}

func TestParseEntersTappedAsCopy(t *testing.T) {
	effect := entersAsCopyEffect(t, "Vesuva",
		"You may have this land enter tapped as a copy of any land on the battlefield.")
	if !effect.EntersAsCopyOptional {
		t.Error("expected optional copy")
	}
	if !effect.EntersAsCopyTapped {
		t.Error("expected enters-tapped rider")
	}
	if len(effect.Selection.RequiredTypesAny) != 1 || effect.Selection.RequiredTypesAny[0] != CardTypeLand {
		t.Errorf("selection required types = %v, want [Land]", effect.Selection.RequiredTypesAny)
	}
}

func TestParseEntersAsCopyNotTappedByDefault(t *testing.T) {
	effect := entersAsCopyEffect(t, "Clone",
		"You may have Clone enter the battlefield as a copy of any creature on the battlefield.")
	if effect.EntersAsCopyTapped {
		t.Error("plain enters-as-copy must not set the enters-tapped rider")
	}
}

func TestParseEntersAsCopyGrantedAbilityRider(t *testing.T) {
	effect := entersAsCopyEffect(t, "Estrid's Invocation",
		"You may have this enchantment enter as a copy of an enchantment you control, "+
			"except it has \"At the beginning of your upkeep, you may exile this enchantment. "+
			"If you do, return it to the battlefield under its owner's control.\"")
	if !effect.EntersAsCopyOptional {
		t.Error("expected optional copy")
	}
	if !effect.EntersAsCopyGrantedAbilityRider {
		t.Fatal("expected granted-ability rider marker")
	}
	if effect.EntersAsCopyGrantedAbility == nil {
		t.Fatal("granted ability was not bound by the attach pass")
	}
	inner, diags := effect.EntersAsCopyGrantedAbility.Inner()
	if len(diags) != 0 {
		t.Fatalf("granted ability inner parse diagnostics: %#v", diags)
	}
	if len(inner.Abilities) != 1 || inner.Abilities[0].Kind != AbilityTriggered {
		t.Fatalf("granted ability = %#v, want a single triggered ability", inner.Abilities)
	}
	if effect.Selection.Controller != SelectionControllerYou {
		t.Errorf("controller = %v, want you", effect.Selection.Controller)
	}
	if len(effect.Selection.RequiredTypesAny) != 1 || effect.Selection.RequiredTypesAny[0] != CardTypeEnchantment {
		t.Errorf("required types = %v, want [enchantment]", effect.Selection.RequiredTypesAny)
	}
}

func TestParseEntersAsCopyGrantedAbilityRequiresQuotedAbility(t *testing.T) {
	// A bare "except it has" with no quoted ability must fail closed rather than
	// recognizing a granted-ability rider with nothing bound.
	doc, _ := Parse("You may have this enchantment enter as a copy of an enchantment you control, except it has.",
		Context{CardName: "Not Estrid"})
	for a := range doc.Abilities {
		for s := range doc.Abilities[a].Sentences {
			for _, effect := range doc.Abilities[a].Sentences[s].Effects {
				if effect.EntersAsCopy {
					t.Fatalf("bare \"except it has\" must not parse as an enters-as-copy effect: %#v", effect)
				}
			}
		}
	}
}
