package parser

import (
	"testing"

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
		{"Vesuva", "You may have this land enter tapped as a copy of any land on the battlefield."},
		{"Synth Infiltrator", "You may have this creature enter as a copy of any creature on the battlefield, except it's a Synth artifact in addition to its other types."},
	}
	for _, tc := range cases {
		if hasEntersAsCopyEffect(t, tc.name, tc.text) {
			t.Errorf("%s: expected enters-as-copy to fail closed", tc.name)
		}
	}
}
