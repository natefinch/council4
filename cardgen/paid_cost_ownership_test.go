package cardgen

import "testing"

func TestPaidCostAmbiguousSpellOwnershipRefusesWholeCard(t *testing.T) {
	for _, text := range []string{
		"As an additional cost to cast this spell, sacrifice a creature.\nSacrifice another creature.\nIf the sacrificed creature was red, draw a card.",
		"As an additional cost to cast this spell, discard a card.\nDiscard another card.\nIf the discarded card was a land card, draw a card.",
		"As an additional cost to cast this spell, sacrifice a creature.\nSacrifice another creature.\nIf the sacrificed creature was not red, draw a card.",
		"As an additional cost to cast this spell, sacrifice a creature.\nYou may sacrifice a creature rather than pay this spell's mana cost.\nIf the sacrificed creature was red, draw a card.",
		"As an additional cost to cast this spell, sacrifice a creature.\nIf the sacrificed creature was red, draw a card.\nYou may sacrifice a creature rather than pay this spell's mana cost.",
		"As an additional cost to cast this spell, discard a card.\nYou may discard a card rather than pay this spell's mana cost.\nIf the discarded card was a land card, draw a card.",
		"As an additional cost to cast this spell, sacrifice a creature.\nYou may sacrifice a creature rather than pay this spell's mana cost.\nDraw a card. If the sacrificed creature had toughness 4 or greater, draw two cards instead.",
		"As an additional cost to cast this spell, sacrifice a creature.\nIf the sacrificed creature was red, draw a card.\nAs an additional cost to cast this spell, sacrifice an artifact.",
	} {
		t.Run(text, func(t *testing.T) {
			assertCardUnsupported(t, &ScryfallCard{Name: "Ambiguous Spell Payment", Layout: "normal",
				ManaCost: "{2}", TypeLine: "Sorcery", OracleText: text})
		})
	}
}

func TestPaidCostDifferentAlternativeDomainRetainsExactBinding(t *testing.T) {
	card := &ScryfallCard{Name: "Different Spell Payments", Layout: "normal", ManaCost: "{2}", TypeLine: "Sorcery",
		OracleText: "As an additional cost to cast this spell, discard a card.\nYou may sacrifice a creature rather than pay this spell's mana cost.\nIf the discarded card was a land card, draw a card."}
	assertCardPaths(t, card, "Object.Val.kind = game.ObjectReferencePaidCost",
		"Object.Val.costKey = \"paid-cost-1-1\"", "AdditionalCosts[0].SubjectKey = \"paid-cost-1-1\"")
}
