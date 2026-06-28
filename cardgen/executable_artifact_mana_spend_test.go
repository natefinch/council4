package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableArtifactManaSpendRiders covers the artifact mana-spend
// restriction family: each card's tap mana ability produces mana carrying a
// restriction-only spend rider tagged with the recognized closed condition.
func TestGenerateExecutableArtifactManaSpendRiders(t *testing.T) {
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
			name:      "Castle Doom",
			typeLine:  "Land",
			oracle:    "{T}: Add {C}.\n{T}: Add one mana of any color. Spend this mana only to cast an artifact spell.",
			condition: "game.ManaSpendCastArtifactSpellOnly,",
		},
		{
			name:      "Power Depot",
			typeLine:  "Land",
			oracle:    "This land enters tapped.\n{T}: Add {C}.\n{T}: Add one mana of any color. Spend this mana only to cast artifact spells or activate abilities of artifacts.",
			condition: "game.ManaSpendCastOrActivateArtifact,",
		},
		{
			name:      "Soldevi Machinist",
			typeLine:  "Artifact Creature — Construct",
			oracle:    "{T}: Add {C}{C}. Spend this mana only to activate abilities of artifacts.",
			power:     "0",
			toughness: "4",
			condition: "game.ManaSpendActivateArtifactAbility,",
		},
		{
			name:      "Guidelight Optimizer",
			typeLine:  "Artifact Creature — Construct",
			oracle:    "{T}: Add {C}. Spend this mana only to cast an artifact spell or activate an ability.",
			power:     "2",
			toughness: "3",
			condition: "game.ManaSpendCastArtifactOrActivateAbility,",
		},
		{
			name:      "Vedalken Engineer",
			typeLine:  "Creature — Vedalken Artificer",
			oracle:    "{T}: Add two mana of any one color. Spend this mana only to cast artifact spells or activate abilities of artifacts.",
			power:     "0",
			toughness: "3",
			condition: "game.ManaSpendCastOrActivateArtifact,",
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
