package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableWormfangExileHandReturn exercises the linked
// entire-hand exile-and-return-to-hand mechanism (issue #1486): an
// enters-the-battlefield clause that exiles the controller's whole hand under a
// source-keyed link, paired with a leaves-the-battlefield clause that returns
// exactly that exiled set to its owners' hands. Both halves must lower under the
// constant exile-hand-return key so the runtime binding pairs them.
func TestGenerateExecutableWormfangExileHandReturn(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Wormfang Behemoth",
		Layout:   "normal",
		ManaCost: "{5}{U}",
		TypeLine: "Creature — Beast",
		OracleText: "When Wormfang Behemoth enters, exile all cards from your hand.\n" +
			"When Wormfang Behemoth leaves the battlefield, return the exiled cards to their owner's hand.",
		Colors: []string{"U"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "o")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.ExileEntireHand",
		"Player:    game.ControllerReference(),",
		`LinkedKey: game.LinkedKey("exile-hand-return"),`,
		"Primitive: game.ReturnExiledCardsToHand",
		"game.EventPermanentEnteredBattlefield",
		"game.EventZoneChanged",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
