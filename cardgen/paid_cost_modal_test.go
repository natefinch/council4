package cardgen

import "testing"

func TestPaidCostModalEffectAmbiguityRefusesWholeCard(t *testing.T) {
	for _, text := range []string{
		"As an additional cost to cast this spell, sacrifice a creature.\nChoose one or both \u2014\n\u2022 Sacrifice another creature.\n\u2022 If the sacrificed creature was red, draw a card.",
		"As an additional cost to cast this spell, discard a card.\nChoose one or both \u2014\n\u2022 Discard a card.\n\u2022 If the discarded card was a land card, draw a card.",
		"As an additional cost to cast this spell, sacrifice a creature.\nChoose one or both \u2014\n\u2022 Sacrifice another creature.\n\u2022 If a Saproling was sacrificed this way, draw a card.",
	} {
		assertCardUnsupported(t, &ScryfallCard{Name: "Ambiguous Modal Payment", Layout: "normal", TypeLine: "Sorcery", OracleText: text})
	}
}

func TestPaidCostExclusiveModeRetainsExactBinding(t *testing.T) {
	card := &ScryfallCard{Name: "Exclusive Modal Payment", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "As an additional cost to cast this spell, sacrifice a creature.\nChoose one \u2014\n\u2022 Sacrifice another creature.\n\u2022 If the sacrificed creature was red, draw a card."}
	assertCardPaths(t, card, "Object.Val.kind = game.ObjectReferencePaidCost",
		"Object.Val.costKey = \"paid-cost-1-1\"", "AdditionalCosts[0].SubjectKey = \"paid-cost-1-1\"")
}
