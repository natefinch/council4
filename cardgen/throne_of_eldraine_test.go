package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceThroneOfEldraine(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Throne of Eldraine",
		Layout:   "normal",
		ManaCost: "{5}",
		TypeLine: "Legendary Artifact",
		OracleText: "As Throne of Eldraine enters, choose a color.\n" +
			"{T}: Add four mana of the chosen color. Spend this mana only to cast monocolored spells of that color.\n" +
			"{3}, {T}: Draw two cards. Spend only mana of the chosen color to activate this ability.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		// As-enters color choice.
		"game.EntryColorChoiceReplacement",
		// Mana ability adds four of the chosen color.
		"game.ManaAbility{",
		"Amount:          game.Fixed(4)",
		`EntryChoiceFrom: game.ChoiceKey("oracle-entry-color")`,
		// Monocolored-chosen-color spend rider on the produced mana.
		"SpendRider: opt.Val(game.ManaSpendRider{",
		"Condition:   game.ManaSpendCastMonocoloredSpellOfChosenColor",
		"Restriction: game.ManaSpendRestrictedToCondition",
		// Activated draw ability restricted to chosen-color mana.
		"game.ActivatedAbility{",
		"game.Draw{",
		"ManaCostRestrictedToEntryChosenColor: true",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
