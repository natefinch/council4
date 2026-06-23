package cardgen

import (
	"strings"
	"testing"
)

func eyeOfUginCard() *ScryfallCard {
	return &ScryfallCard{
		Name:     "Eye of Ugin",
		Layout:   "normal",
		ManaCost: "",
		TypeLine: "Legendary Land",
		OracleText: "Colorless Eldrazi spells you cast cost {2} less to cast.\n" +
			"{7}, {T}: Search your library for a colorless creature card, reveal it, put it into your hand, then shuffle.",
	}
}

// TestGenerateExecutableCardSourceEyeOfUgin asserts the activated ability lowers
// a search for a colorless creature card, exercising the colorless search-filter
// support independently of the optional-sacrifice flow.
func TestGenerateExecutableCardSourceEyeOfUgin(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(eyeOfUginCard(), "e")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.Search{",
		"Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}, Colorless: true}",
		"Reveal:      true",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
