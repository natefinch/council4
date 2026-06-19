package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceSmotheringTithe(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Smothering Tithe",
		Layout:     "normal",
		ManaCost:   "{3}{W}",
		TypeLine:   "Enchantment",
		OracleText: `Whenever an opponent draws a card, that player may pay {2}. If the player doesn't, you create a Treasure token. (It's an artifact with "{T}, Sacrifice this token: Add one mana of any color.")`,
	}, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.EventCardDrawn",
		"game.TriggerPlayerOpponent",
		"Primitive: game.Pay",
		"opt.Val(game.EventPlayerReference())",
		"cost.O(2)",
		"PublishResult: game.ResultKey(\"unless-paid\")",
		"Primitive: game.CreateToken",
		"Succeeded: game.TriFalse",
		"Subtypes: []types.Sub{types.Treasure}",
		"ManaAbilities: []game.ManaAbility",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
	if strings.Contains(source, "Optional: true") {
		t.Fatalf("mandatory failure consequence rendered optional:\n%s", source)
	}
}
