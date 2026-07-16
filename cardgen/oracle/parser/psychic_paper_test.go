package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

const psychicPaperOracle = "As this Equipment becomes attached to a creature, choose a creature card name and a creature type.\n" +
	"Equipped creature has ward {1}, it can't be blocked, and its name and creature type are the last chosen name and creature type.\n" +
	"Equip {2}"

func TestParsePsychicPaperAttachmentChoicesAndIdentity(t *testing.T) {
	document, diagnostics := Parse(psychicPaperOracle, Context{CardName: "Psychic Paper"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 3 {
		t.Fatalf("abilities = %d, want 3", len(document.Abilities))
	}
	choice := document.Abilities[0]
	if choice.Kind != AbilityReplacement || choice.AttachmentChoices == nil {
		t.Fatalf("attachment choice ability = %#v", choice)
	}
	if choice.AttachmentChoices.CardNameType != types.Creature ||
		choice.AttachmentChoices.SubtypeOfType != types.Creature {
		t.Fatalf("attachment choices = %#v", choice.AttachmentChoices)
	}
	statics := document.Abilities[1].StaticDeclarations
	if len(statics) != 3 {
		t.Fatalf("static declarations = %#v, want ward, unblockable, and identity", statics)
	}
	if statics[0].Kind != StaticDeclarationKeywordGrant ||
		statics[1].Kind != StaticDeclarationRule ||
		statics[2].Kind != StaticDeclarationContinuousAttachmentChoiceIdentity ||
		statics[2].ChoiceCardType != types.Creature {
		t.Fatalf("static declarations = %#v", statics)
	}
}

func TestParsePsychicPaperGrammarFailsClosedOnNeighbors(t *testing.T) {
	for _, oracle := range []string{
		"As this Equipment becomes attached to a creature, choose a card name and a creature type.",
		"As this Equipment becomes attached to a creature, choose a creature card name or a creature type.",
		"Equipped creature has ward {1}, it can't be blocked, and its name and creature type are a chosen name and creature type.",
	} {
		document, _ := Parse(oracle, Context{CardName: "Test Equipment"})
		if len(document.Abilities) != 1 {
			t.Fatalf("Parse(%q) abilities = %d", oracle, len(document.Abilities))
		}
		ability := document.Abilities[0]
		if ability.AttachmentChoices != nil {
			t.Fatalf("Parse(%q) produced attachment choices %#v", oracle, ability.AttachmentChoices)
		}
		for _, declaration := range ability.StaticDeclarations {
			if declaration.Kind == StaticDeclarationContinuousAttachmentChoiceIdentity {
				t.Fatalf("Parse(%q) produced attachment identity %#v", oracle, declaration)
			}
		}
	}
}
