package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableExplicitORingReturn(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		manaCost   string
		oracleText string
		power      *string
		toughness  *string
	}{
		{
			name:     "Journey to Nowhere",
			typeLine: "Enchantment",
			manaCost: "{1}{W}",
			oracleText: "When Journey to Nowhere enters, exile target creature.\n" +
				"When Journey to Nowhere leaves the battlefield, return the exiled card to the battlefield under its owner's control.",
		},
		{
			name:     "Oblivion Ring",
			typeLine: "Enchantment",
			manaCost: "{2}{W}",
			oracleText: "When Oblivion Ring enters, exile another target nonland permanent.\n" +
				"When Oblivion Ring leaves the battlefield, return the exiled card to the battlefield under its owner's control.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       test.name,
				Layout:     "normal",
				ManaCost:   test.manaCost,
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
				Colors:     []string{"W"},
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "o")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, wanted := range []string{
				"Primitive: game.Exile",
				"Object:         game.TargetPermanentReference(0)",
				`ExileLinkedKey: game.LinkedKey("exile-until-leaves")`,
				"Primitive: game.PutOnBattlefield",
				`game.LinkedBattlefieldSource(game.LinkedKey("exile-until-leaves"))`,
				"game.EventZoneChanged",
			} {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}
