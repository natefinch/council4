package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

func TestAttachmentChoicesReplacementCarriesTypedChoices(t *testing.T) {
	ability := AttachmentChoicesReplacement("choose", types.Creature, types.Creature)
	if ability.Replacement.AttachCardNameChoiceType != types.Creature ||
		ability.Replacement.AttachSubtypeChoiceType != types.Creature {
		t.Fatalf("replacement = %#v", ability.Replacement)
	}
}

func TestClonePreservesPsychicPaperChoices(t *testing.T) {
	g := NewGame([NumPlayers]PlayerConfig{})
	g.CardNameCatalog = map[types.Card][]string{
		types.Creature: {"Alpha Creature"},
	}
	permanent := &Permanent{
		ObjectID:   g.IDGen.Next(),
		Owner:      Player1,
		Controller: Player1,
		EntryChoices: map[ChoiceKey]ResolutionChoiceResult{
			AttachmentCardNameChoiceKey: {
				Kind:     ResolutionChoiceCardName,
				CardName: "Alpha Creature",
			},
			AttachmentSubtypeChoiceKey: {
				Kind:    ResolutionChoiceSubtype,
				Subtype: types.Elf,
			},
		},
	}
	g.Battlefield = append(g.Battlefield, permanent)

	clone := g.Clone()
	cloned := clone.Battlefield[0]
	if cloned.EntryChoices[AttachmentCardNameChoiceKey].CardName != "Alpha Creature" ||
		cloned.EntryChoices[AttachmentSubtypeChoiceKey].Subtype != types.Elf {
		t.Fatalf("cloned choices = %#v", cloned.EntryChoices)
	}
	delete(cloned.EntryChoices, AttachmentCardNameChoiceKey)
	if _, ok := permanent.EntryChoices[AttachmentCardNameChoiceKey]; !ok {
		t.Fatal("clone choice map aliases original")
	}
	clone.CardNameCatalog[types.Creature][0] = "Changed"
	if g.CardNameCatalog[types.Creature][0] != "Alpha Creature" {
		t.Fatal("clone card-name catalog aliases original")
	}
}
