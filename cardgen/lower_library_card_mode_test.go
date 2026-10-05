package cardgen

import "testing"

func TestLibraryCardSelectedModesOwnTheirProductNamespaces(t *testing.T) {
	card := &ScryfallCard{
		Name: "Independent Observation Modes", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "Choose two \u2014\n" +
			"\u2022 Reveal the top card of your library. If it's a creature card, you gain 1 life.\n" +
			"\u2022 If you control a creature, reveal the top card of your library. If it's a creature card, you gain 2 life.",
	}
	assertCardPaths(t, card,
		"SpellAbility.Val.MaxModes = 2",
		"Modes[0].Sequence[0].LocalProducts.Links[0] = \"sequence-effect-0-product\"",
		"Modes[1].Sequence[0].LocalProducts.Links[0] = \"sequence-effect-0-product\"",
		"Modes[1].Sequence[0].Condition.Exists = true",
		"Modes[1].Sequence[1].Condition.Val.Condition.Val.Object.Val.linkID = \"sequence-effect-0-product\"",
	)
}
