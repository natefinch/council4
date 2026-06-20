package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceEnteringTriggerMultiplierCategory(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		card  *ScryfallCard
		wants []string
	}{
		"panharmonicon": {
			card: &ScryfallCard{
				Name:       "Panharmonicon",
				Layout:     "normal",
				ManaCost:   "{4}",
				TypeLine:   "Artifact",
				OracleText: "If an artifact or creature entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.",
			},
			wants: []string{
				"game.RuleEffectAdditionalTriggerForEnteringPermanent",
				"types.Card{types.Artifact, types.Creature}",
			},
		},
		"yarok": {
			card: &ScryfallCard{
				Name:       "Yarok, the Desecrated",
				Layout:     "normal",
				ManaCost:   "{1}{B}{G}{U}",
				TypeLine:   "Legendary Creature — Horror",
				OracleText: "Deathtouch, lifelink\nIf a permanent entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.",
				Power:      new("3"),
				Toughness:  new("5"),
			},
			wants: []string{
				"game.RuleEffectAdditionalTriggerForEnteringPermanent",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(tc.card, "p")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, wanted := range tc.wants {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

func TestGenerateExecutableCardSourceEnteringTriggerMultiplierFailsClosed(t *testing.T) {
	t.Parallel()
	for name, ability := range map[string]string{
		"subtype filter": "If a Wizard you control entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.",
		"twice wording":  "If a permanent entering causes a triggered ability of a permanent you control to trigger, that ability triggers twice.",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Near Miss Harmonicon",
				Layout:     "normal",
				ManaCost:   "{4}",
				TypeLine:   "Artifact",
				OracleText: ability,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "n")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" || len(diagnostics) == 0 {
				t.Fatalf("source = %q, diagnostics = %#v; want fail closed", source, diagnostics)
			}
		})
	}
}
