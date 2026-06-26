package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceManaProductionMultiplier(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		card  *ScryfallCard
		wants []string
	}{
		"mana reflection": {
			card: &ScryfallCard{
				Name:       "Mana Reflection",
				Layout:     "normal",
				ManaCost:   "{4}{G}{G}",
				TypeLine:   "Enchantment",
				OracleText: "If you tap a permanent for mana, it produces twice as much of that mana instead.",
			},
			wants: []string{
				"game.RuleEffectManaProductionMultiplier",
				"ManaProductionMultiplier: 2",
			},
		},
		"nyxbloom ancient": {
			card: &ScryfallCard{
				Name:       "Nyxbloom Ancient",
				Layout:     "normal",
				ManaCost:   "{4}{G}{G}{G}",
				TypeLine:   "Enchantment Creature — Elemental",
				OracleText: "Trample\nIf you tap a permanent for mana, it produces three times as much of that mana instead.",
				Power:      new("5"),
				Toughness:  new("5"),
			},
			wants: []string{
				"game.RuleEffectManaProductionMultiplier",
				"ManaProductionMultiplier: 3",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(tc.card, "n")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			normalized := strings.Join(strings.Fields(source), " ")
			for _, wanted := range tc.wants {
				if !strings.Contains(normalized, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}
