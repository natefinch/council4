package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceEDHStatics(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		card  *ScryfallCard
		wants []string
	}{
		"leyline of anticipation": {
			card: &ScryfallCard{
				Name:       "Leyline of Anticipation",
				Layout:     "normal",
				ManaCost:   "{2}{U}{U}",
				TypeLine:   "Enchantment",
				OracleText: "If this card is in your opening hand, you may begin the game with it on the battlefield.\nYou may cast spells as though they had flash.",
			},
			wants: []string{
				"game.RuleEffectCastSpellsAsThoughFlash",
			},
		},
		"elesh norn mother of machines": {
			card: &ScryfallCard{
				Name:       "Elesh Norn, Mother of Machines",
				Layout:     "normal",
				ManaCost:   "{4}{W}",
				TypeLine:   "Legendary Creature — Phyrexian Praetor",
				OracleText: "Vigilance\nIf a permanent entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.\nPermanents entering don't cause abilities of permanents your opponents control to trigger.",
				Power:      new("4"),
				Toughness:  new("7"),
			},
			wants: []string{
				"game.RuleEffectAdditionalTriggerForEnteringPermanent",
				"game.RuleEffectSuppressOpponentEnteringTriggers",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(tc.card, "e")
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
