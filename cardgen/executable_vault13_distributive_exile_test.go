package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableVault13DistributiveExile exercises the per-player
// distributive exile-until-leaves mechanism (issue #1486): a Saga chapter that
// exiles up to one permanent each player controls under a source-keyed link,
// paired with a later chapter that returns a fixed subset of that linked set to
// the battlefield and bottoms the rest, plus the face-level synthesized
// leaves-the-battlefield safety net that releases the captives if the Saga
// leaves before the payoff chapter. The exile and both returns must read the
// same exile-until-leaves key so the runtime binding pairs them.
func TestGenerateExecutableVault13DistributiveExile(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Vault 13: Dweller's Journey",
		Layout:   "saga",
		ManaCost: "{3}{G}",
		TypeLine: "Enchantment — Saga",
		OracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)\n" +
			"I — For each player, exile up to one other target enchantment or creature that player controls until this Saga leaves the battlefield.\n" +
			"II — You gain 2 life and scry 2.\n" +
			"III — Return two cards exiled with this Saga to the battlefield under their owners' control and put the rest on the bottom of their owners' libraries.",
		Colors: []string{"G"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "o")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.ExileForEachPlayer{",
		"Chooser:   game.ControllerReference(),",
		"Selection: game.Selection{RequiredTypesAny: []types.Card{types.Enchantment, types.Creature}, ExcludeSource: true},",
		`LinkedKey: game.LinkedKey("exile-until-leaves"),`,
		"Primitive: game.ReturnLinkedExiledCardsToBattlefield{",
		"Amount:              game.Fixed(2),",
		"RestToLibraryBottom: true,",
		"Primitive: game.PutOnBattlefield{",
		`Source: game.LinkedBattlefieldSource(game.LinkedKey("exile-until-leaves")),`,
		"game.EventZoneChanged",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
