package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceHeraldicBanner(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Heraldic Banner",
		Layout:     "normal",
		ManaCost:   "{3}",
		TypeLine:   "Artifact",
		OracleText: "As Heraldic Banner enters, choose a color.\nCreatures you control of the chosen color get +1/+0.\n{T}: Add one mana of the chosen color.",
	}, "h")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.EntryColorChoiceReplacement",
		"game.ObjectControlledGroup(game.SourcePermanentReference()",
		"ColorChoice: game.ColorChoiceSourceEntry",
		"PowerDelta: 1",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
