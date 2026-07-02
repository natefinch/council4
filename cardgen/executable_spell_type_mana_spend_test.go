package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableSpellTypeManaSpendRiders covers the spell-type mana-spend
// restriction family: each card's tap mana ability produces mana carrying a
// restriction-only spend rider tagged with the recognized closed condition,
// across colorless, any-one-color, any-color, and typed mana shapes.
func TestGenerateExecutableSpellTypeManaSpendRiders(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		typeLine  string
		oracle    string
		power     string
		toughness string
		condition string
	}{
		{
			name:      "Vodalian Arcanist",
			typeLine:  "Creature — Merfolk Wizard",
			oracle:    "{T}: Add {C}. Spend this mana only to cast an instant or sorcery spell.",
			power:     "1",
			toughness: "3",
			condition: "game.ManaSpendCastInstantOrSorcerySpell,",
		},
		{
			name:      "Cormela, Glamour Thief",
			typeLine:  "Legendary Creature — Human Wizard",
			oracle:    "{T}: Add {B}{R}. Spend this mana only to cast instant and/or sorcery spells.",
			power:     "3",
			toughness: "3",
			condition: "game.ManaSpendCastInstantOrSorcerySpell,",
		},
		{
			name:      "Somberwald Sage",
			typeLine:  "Creature — Human Shaman",
			oracle:    "{T}: Add three mana of any one color. Spend this mana only to cast creature spells.",
			power:     "1",
			toughness: "2",
			condition: "game.ManaSpendCastCreatureSpell,",
		},
		{
			name:      "Pillar of the Paruns",
			typeLine:  "Land",
			oracle:    "{T}: Add one mana of any color. Spend this mana only to cast a multicolored spell.",
			condition: "game.ManaSpendCastMulticoloredSpell,",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       tc.name,
				Layout:     "normal",
				TypeLine:   tc.typeLine,
				OracleText: tc.oracle,
			}
			if tc.power != "" {
				card.Power = &tc.power
				card.Toughness = &tc.toughness
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "z")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, wanted := range []string{
				"SpendRider: opt.Val(game.ManaSpendRider{",
				tc.condition,
				"Restriction: game.ManaSpendRestrictedToCondition,",
			} {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}
