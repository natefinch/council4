package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableForetellKeyword exercises the Foretell keyword
// (CR 702.144): a card with "Foretell <cost>" lowers to a game.ForetellKeyword
// static ability carrying the foretell mana cost, alongside its normal spell
// ability.
func TestGenerateExecutableForetellKeyword(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Behold the Multiverse",
		Layout:   "normal",
		ManaCost: "{3}{U}",
		TypeLine: "Instant",
		OracleText: "Scry 2, then draw two cards.\n" +
			"Foretell {1}{U} (During your turn, you may pay {2} and exile this card from your hand face down. Cast it on a later turn for its foretell cost.)",
		Colors: []string{"U"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "b")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.ForetellKeyword{Cost:") {
		t.Fatalf("source missing game.ForetellKeyword:\n%s", source)
	}
}
