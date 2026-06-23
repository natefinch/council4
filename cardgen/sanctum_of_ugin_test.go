package cardgen

import (
	"strings"
	"testing"
)

func sanctumOfUginCard() *ScryfallCard {
	return &ScryfallCard{
		Name:     "Sanctum of Ugin",
		Layout:   "normal",
		ManaCost: "",
		TypeLine: "Land",
		OracleText: "{T}: Add {C}.\n" +
			"Whenever you cast a colorless spell with mana value 7 or greater, you may sacrifice this land. If you do, search your library for a colorless creature card, reveal it, put it into your hand, then shuffle.",
	}
}

// TestGenerateExecutableCardSourceSanctumOfUgin asserts the colorless-spell-cast
// trigger lowers both the "colorless" cast condition and the optional sacrifice
// gating a search for a colorless creature card.
func TestGenerateExecutableCardSourceSanctumOfUgin(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(sanctumOfUginCard(), "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"CardSelection: game.Selection{Colorless: true",
		"game.Sacrifice{",
		"Optional:      true",
		"PublishResult: game.ResultKey(\"if-you-do\")",
		"Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}, Colorless: true}",
		"Succeeded: game.TriTrue",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
