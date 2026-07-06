package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutablePlotKeyword exercises the Plot keyword (CR 718): a card
// with "Plot <cost>" lowers to a game.PlotKeyword static ability carrying the
// plot mana cost, alongside its normal spell ability.
func TestGenerateExecutablePlotKeyword(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Plan the Heist",
		Layout:   "normal",
		ManaCost: "{2}{U}",
		TypeLine: "Sorcery",
		OracleText: "Draw two cards.\n" +
			"Plot {1}{U} (You may pay {1}{U} and exile this card from your hand. Cast it as a sorcery on a later turn without paying its mana cost. Plot only as a sorcery.)",
		Colors: []string{"U"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "p")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.PlotKeyword{Cost:") {
		t.Fatalf("source missing game.PlotKeyword:\n%s", source)
	}
}
