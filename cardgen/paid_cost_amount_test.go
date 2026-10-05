package cardgen

import "testing"

func TestPaidCostPredicatesPreserveLegacyCharacteristicAmounts(t *testing.T) {
	for _, text := range []string{
		"Sacrifice a creature: You gain life equal to the sacrificed creature's power.",
		"Sacrifice a creature: You gain life equal to the sacrificed creature's toughness.",
		"Sacrifice a creature: Target player mills cards equal to the sacrificed creature's mana value.",
		"Sacrifice a creature: Draw cards equal to the sacrificed creature's power.",
	} {
		t.Run(text, func(t *testing.T) {
			card := &ScryfallCard{Name: "Legacy Paid Amount", Layout: "normal", TypeLine: "Artifact", OracleText: text}
			assertCardPaths(t, card, "kind = game.ObjectReferenceSacrificedCost")
			assertCardPathsAbsent(t, card, "kind = game.ObjectReferencePaidCost", "SubjectKey")
		})
	}
}

func TestPaidCostAmountAndPredicateComposeIndependently(t *testing.T) {
	card := &ScryfallCard{Name: "Two Paid Consumers", Layout: "normal", TypeLine: "Artifact",
		OracleText: "Sacrifice a creature: You gain life equal to the sacrificed creature's power. If the sacrificed creature had toughness 4 or greater, draw a card."}
	assertCardPaths(t, card, "kind = game.ObjectReferenceSacrificedCost",
		"Object.Val.kind = game.ObjectReferencePaidCost", "Object.Val.costKey = \"paid-cost-1-1\"")
}
