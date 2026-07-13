package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceCagedSun(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Caged Sun",
		Layout:   "normal",
		ManaCost: "{6}",
		TypeLine: "Artifact",
		OracleText: "As Caged Sun enters, choose a color.\n" +
			"Creatures you control of the chosen color get +1/+1.\n" +
			"Whenever a land's ability causes you to add one or more mana of the chosen color, add an additional one mana of that color.",
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		// As-enters color choice.
		"game.EntryColorChoiceReplacement",
		// Dynamic chosen-color group for creatures you control, +1/+1.
		"game.ObjectControlledGroup(game.SourcePermanentReference()",
		"RequiredTypes: []types.Card{types.Creature}",
		"ColorChoice: game.ColorChoiceSourceEntry",
		"PowerDelta:     1",
		"ToughnessDelta: 1",
		// Chosen-color land-mana trigger.
		"game.EventPermanentTapped",
		"RequireTappedForMana:                    true",
		"RequireProducedManaColorFromEntryChoice: true",
		"RequiredTypes: []types.Card{types.Land}",
		// Additional mana of the chosen color routed through the entry choice.
		"game.AddMana{",
		`EntryChoiceFrom: game.ChoiceKey("oracle-entry-color")`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
