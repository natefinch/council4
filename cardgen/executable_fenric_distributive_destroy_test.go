package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableFenricDistributiveDestroy exercises the per-player
// distributive destroy mechanism (issue #1520): a Saga chapter that destroys up
// to one creature each player controls under a source-keyed link, paired with a
// per-controller payoff that creates one token under each destroyed creature's
// last-known controller. The destroy and the token payoff must read the same
// destroyed-for-each-player key so the runtime binding pairs them.
func TestGenerateExecutableFenricDistributiveDestroy(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "The Curse of Fenric",
		Layout:   "saga",
		ManaCost: "{2}{G}{W}",
		TypeLine: "Enchantment — Saga",
		OracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)\n" +
			"I — For each player, destroy up to one target creature that player controls. For each creature destroyed this way, its controller creates a 3/3 green Mutant creature token with deathtouch.\n" +
			"II — Target nontoken creature becomes a 6/6 legendary Horror creature named Fenric and loses all abilities.\n" +
			"III — Target Mutant fights another target creature named Fenric.",
		Colors: []string{"G", "W"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "o")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.DestroyForEachPlayer{",
		"Chooser:   game.ControllerReference(),",
		"Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},",
		`LinkedKey: game.LinkedKey("destroyed-for-each-player"),`,
		"Primitive: game.CreateTokenForEachDestroyed{",
		"Source:    game.TokenDef(",
		"game.DeathtouchStaticBody,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
