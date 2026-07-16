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
				"BeginsGameOnBattlefield: true",
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
		"sphere of safety": {
			card: &ScryfallCard{
				Name:       "Sphere of Safety",
				Layout:     "normal",
				ManaCost:   "{4}{W}",
				TypeLine:   "Enchantment",
				OracleText: "Creatures can't attack you or planeswalkers you control unless their controller pays {X} for each of those creatures, where X is the number of enchantments you control.",
			},
			wants: []string{
				"game.RuleEffectAttackTaxPerCreature",
				"AffectedPlayer: game.PlayerYou",
				"AttackTaxIncludesPlaneswalkers: true",
				"RequiredTypes: []types.Card{types.Enchantment}",
				"Controller: game.ControllerYou",
			},
		},
		"baird steward of argive": {
			card: &ScryfallCard{
				Name:       "Baird, Steward of Argive",
				Layout:     "normal",
				ManaCost:   "{3}{W}",
				TypeLine:   "Legendary Creature — Human Soldier",
				OracleText: "Vigilance\nCreatures can't attack you or planeswalkers you control unless their controller pays {1} for each of those creatures.",
				Power:      new("2"),
				Toughness:  new("4"),
			},
			wants: []string{
				"game.RuleEffectAttackTaxPerCreature",
				"AffectedPlayer: game.PlayerYou",
				"AttackTaxIncludesPlaneswalkers: true",
				"AttackTaxGeneric: 1",
			},
		},
		"collective restraint": {
			card: &ScryfallCard{
				Name:       "Collective Restraint",
				Layout:     "normal",
				ManaCost:   "{3}{U}",
				TypeLine:   "Enchantment",
				OracleText: "Domain — Creatures can't attack you unless their controller pays {X} for each creature they control that's attacking you, where X is the number of basic land types among lands you control.",
			},
			wants: []string{
				"game.RuleEffectAttackTaxPerCreature",
				"AffectedPlayer: game.PlayerYou",
				"AttackTaxScaledAmount: game.AggregateControllerBasicLandTypeCount",
			},
		},
		"opposition agent": {
			card: &ScryfallCard{
				Name:       "Opposition Agent",
				Layout:     "normal",
				ManaCost:   "{2}{B}",
				TypeLine:   "Creature — Human Rogue",
				OracleText: "Flash\nYou control your opponents while they're searching their libraries.\nWhile an opponent is searching their library, they exile each card they find. You may play those cards for as long as they remain exiled, and you may spend mana as though it were mana of any color to cast them.",
				Power:      new("3"),
				Toughness:  new("2"),
			},
			wants: []string{
				"game.RuleEffectControlOpponentSearches",
				"game.RuleEffectExileOpponentSearchFinds",
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
			// Collapse gofmt's column-alignment padding so struct-field
			// assertions don't depend on neighboring field name lengths.
			normalized := strings.Join(strings.Fields(source), " ")
			for _, wanted := range tc.wants {
				if !strings.Contains(normalized, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}
