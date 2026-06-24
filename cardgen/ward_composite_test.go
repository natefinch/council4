package cardgen

import "testing"

// Ward with a composite cost ("Ward—{2}, Pay 2 life.", Captain Howler, Sea
// Scourge): the non-mana payment rides on the Ward static ability through the
// shared activation-cost kernel, mixing mana with an additional pay-life cost.
func TestGenerateWardCompositeManaAndLifeCost(t *testing.T) {
	t.Parallel()
	power, toughness := "5", "4"
	card := &ScryfallCard{
		Name:       "Test Ward Composite",
		Layout:     "normal",
		ManaCost:   "{2}{U}{R}",
		TypeLine:   "Creature — Shark Pirate",
		Power:      &power,
		Toughness:  &toughness,
		OracleText: "Ward—{2}, Pay 2 life.",
	}
	generatedSourceContains(t, card, []string{
		"game.WardStaticAbilityWithCosts(cost.Mana{cost.O(2)}, []cost.Additional{",
		"cost.AdditionalPayLife",
	})
}

// Ward with a non-mana cost ("Ward—Sacrifice a creature."): the Ward static
// ability carries only an additional cost and no mana.
func TestGenerateWardNonManaSacrificeCost(t *testing.T) {
	t.Parallel()
	power, toughness := "3", "3"
	card := &ScryfallCard{
		Name:       "Test Ward Sacrifice",
		Layout:     "normal",
		ManaCost:   "{1}{B}",
		TypeLine:   "Creature — Horror",
		Power:      &power,
		Toughness:  &toughness,
		OracleText: "Ward—Sacrifice a creature.",
	}
	generatedSourceContains(t, card, []string{
		"game.WardStaticAbilityWithCosts(cost.Mana{}, []cost.Additional{",
		"cost.AdditionalSacrifice",
	})
}

// Ward with a pay-life cost ("Ward—Pay 3 life."): the most common non-mana Ward
// form across the corpus.
func TestGenerateWardPayLifeCost(t *testing.T) {
	t.Parallel()
	power, toughness := "2", "2"
	card := &ScryfallCard{
		Name:       "Test Ward Pay Life",
		Layout:     "normal",
		ManaCost:   "{1}{G}",
		TypeLine:   "Creature — Beast",
		Power:      &power,
		Toughness:  &toughness,
		OracleText: "Ward—Pay 3 life.",
	}
	generatedSourceContains(t, card, []string{
		"game.WardStaticAbilityWithCosts(cost.Mana{}, []cost.Additional{",
		"cost.AdditionalPayLife",
	})
}

// A target-creature pump scaled by the triggering discard's card count, lowered
// as one clause of an ordered trigger sequence ("Whenever you discard one or
// more cards, target creature gets +2/+0 until end of turn for each card
// discarded this way. ..."). The triggering-event card count must remain
// available to the sequence clause so the per-card multiplier resolves.
func TestGenerateDiscardScaledPumpInSequence(t *testing.T) {
	t.Parallel()
	power, toughness := "5", "4"
	card := &ScryfallCard{
		Name:      "Test Discard Pump",
		Layout:    "normal",
		ManaCost:  "{2}{U}{R}",
		TypeLine:  "Creature — Shark Pirate",
		Power:     &power,
		Toughness: &toughness,
		OracleText: "Whenever you discard one or more cards, target creature gets +2/+0 " +
			"until end of turn for each card discarded this way. You gain 1 life.",
	}
	generatedSourceContains(t, card, []string{
		"game.EventCardDiscarded",
		"game.DynamicAmountEventCardCount",
		"Multiplier: 2",
	})
}
