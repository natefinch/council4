package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourcePsychicPaper(t *testing.T) {
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Psychic Paper",
		Layout:     "normal",
		ManaCost:   "{2}",
		TypeLine:   "Artifact — Equipment",
		OracleText: "As this Equipment becomes attached to a creature, choose a creature card name and a creature type.\nEquipped creature has ward {1}, it can't be blocked, and its name and creature type are the last chosen name and creature type.\nEquip {2}",
	}, "p")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		`game.AttachmentChoicesReplacement("As this Equipment becomes attached`,
		"SetNameFromSourceChoice: game.AttachmentCardNameChoiceKey",
		"SetSubtypeFromSourceChoice: game.AttachmentSubtypeChoiceKey",
		"SetSubtypeChoiceType:       types.Creature",
		"new(game.WardStaticAbility(cost.Mana{cost.O(1)}))",
		"Kind:             game.RuleEffectCantBeBlocked",
		"game.EquipActivatedAbility(cost.Mana{cost.O(2)})",
	} {
		if !strings.Contains(source, want) {
			t.Errorf("generated source missing %q:\n%s", want, source)
		}
	}
}
