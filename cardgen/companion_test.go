package cardgen

import "testing"

func TestGenerateExecutableCompanionSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Companion Sage",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Whale Hydra",
		OracleText: "Companion — Your starting deck contains only cards with mana value 3 or greater and land cards.\nWhen this creature enters, draw a card.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if source == "" {
		t.Fatal("empty generated source")
	}
}
