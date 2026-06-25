package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableHelvaultNamedToken exercises the named multi-keyword
// legendary token generator (issue #1486): a Saga chapter that creates "Avacyn,
// a legendary 8/8 white Angel creature token with flying, vigilance, and
// indestructible" through the leading "Create <Name>, a ..." form, paired with
// the per-player distributive exile-until-leaves chapter whose "non-Saga,
// nonland permanent" filter joins two negated card types with a comma. The token
// definition must carry the legendary supertype, the explicit name, and one
// static-ability body per keyword.
func TestGenerateExecutableHelvaultNamedToken(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Battle at the Helvault",
		Layout:   "saga",
		ManaCost: "{4}{W}{W}",
		TypeLine: "Enchantment — Saga",
		OracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)\n" +
			"I, II — For each player, exile up to one target non-Saga, nonland permanent that player controls until this Saga leaves the battlefield.\n" +
			"III — Create Avacyn, a legendary 8/8 white Angel creature token with flying, vigilance, and indestructible.",
		Colors: []string{"W"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "b")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.ExileForEachPlayer{",
		"Selection: game.Selection{ExcludedTypes: []types.Card{types.Land}, ExcludedSubtype: types.Sub(\"Saga\")},",
		`LinkedKey: game.LinkedKey("exile-until-leaves"),`,
		"Primitive: game.CreateToken{",
		"Source: game.TokenDef(battleAtTheHelvaultToken),",
		`Name:       "Avacyn",`,
		"Supertypes: []types.Super{types.Legendary},",
		"Subtypes:   []types.Sub{types.Angel},",
		"Power:      opt.Val(game.PT{Value: 8}),",
		"game.FlyingStaticBody,",
		"game.VigilanceStaticBody,",
		"game.IndestructibleStaticBody,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
