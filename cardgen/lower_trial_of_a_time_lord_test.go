package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceTrialOfATimeLord exercises the linked exile
// disposal arm of a vote: Trial of a Time Lord exiles creatures with its early
// chapters under the exile-until-leaves link, then chapter IV votes innocent or
// guilty. The guilty arm "the owner of each card exiled with this Saga puts that
// card on the bottom of their library." lowers to a
// game.PutLinkedExiledCardsInLibrary primitive consuming that link, gated on the
// negative margin that means the second option (guilty) won.
func TestGenerateExecutableCardSourceTrialOfATimeLord(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Trial of a Time Lord",
		Layout:     "normal",
		TypeLine:   "Enchantment — Saga",
		ManaCost:   "{1}{W}{W}",
		OracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after IV.)\nI, II, III — Exile target nontoken creature an opponent controls until this Saga leaves the battlefield.\nIV — Starting with you, each player votes for innocent or guilty. If guilty gets more votes, the owner of each card exiled with this Saga puts that card on the bottom of their library.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.Vote{",
		`"innocent"`,
		`"guilty"`,
		"game.PutLinkedExiledCardsInLibrary{",
		`LinkedKey: game.LinkedKey("exile-until-leaves")`,
		"Bottom:    true,",
		"AmountRange: opt.Val(game.IntRange{Min: -4, Max: -1})",
	} {
		if !strings.Contains(normalizeSource(source), normalizeSource(wanted)) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
