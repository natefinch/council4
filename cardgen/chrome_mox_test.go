package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceChromeMox(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Chrome Mox",
		Layout:     "normal",
		ManaCost:   "{0}",
		TypeLine:   "Artifact",
		OracleText: "Imprint — When this artifact enters, you may exile a nonartifact, nonland card from your hand.\n{T}: Add one mana of any of the exiled card's colors.",
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.EventPermanentEnteredBattlefield",
		"Primitive: game.ChooseFromZone{",
		"Optional: true",
		"PublishLinked:",
		"types.Artifact",
		"types.Land",
		"game.TapLinkedExileColorManaAbility",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
