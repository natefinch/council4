package cardgen

import "testing"

func TestGeneratePrimalGrowth(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Primal Growth",
		Layout:   "normal",
		ManaCost: "{2}{G}",
		TypeLine: "Sorcery",
		OracleText: "Kicker—Sacrifice a creature. (You may sacrifice a creature in addition to any other costs as you cast this spell.)\n" +
			"Search your library for a basic land card, put that card onto the battlefield, then shuffle. If this spell was kicked, instead search your library for up to two basic land cards, put them onto the battlefield, then shuffle.",
	}
	generatedSourceContains(t, card, []string{
		"game.KickerKeyword{Cost: cost.Mana{}, AdditionalCosts: []cost.Additional{",
		"Kind:               cost.AdditionalSacrifice",
		"PermanentType:      types.Creature",
		"Amount: game.Fixed(1)",
		"Amount: game.Fixed(2)",
		"Destination: zone.Battlefield",
		"Supertypes: []types.Super{types.Basic}",
		"Negate:         true",
		"SpellWasKicked: true",
	})
}

func TestConditionalInsteadSearchPreservesDynamicAmountAndDestination(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Dynamic Kicker Search",
		Layout:   "normal",
		ManaCost: "{2}{G}",
		TypeLine: "Sorcery",
		OracleText: "Kicker {G} (You may pay an additional {G} as you cast this spell.)\n" +
			"Search your library for a basic land card, put that card onto the battlefield, then shuffle. If this spell was kicked, instead search your library for up to X basic land cards, where X is the number of lands you control, put those cards onto the battlefield tapped, then shuffle.",
	}
	generatedSourceContains(t, card, []string{
		"Amount: game.Fixed(1)",
		"Amount: game.Dynamic(game.DynamicAmount{",
		"Kind:       game.DynamicAmountCountSelector",
		"EntersTapped: true",
		"SpellWasKicked: true",
	})
}
