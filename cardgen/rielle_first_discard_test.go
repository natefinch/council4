package cardgen

import "testing"

// Rielle, the Everwise: a first-discard-each-turn batch trigger ("Whenever you
// discard one or more cards for the first time each turn, draw that many
// cards") paired with a dynamic self-buff static. The trigger gates on the
// first discard occurrence of the turn (PlayerEventOrdinalThisTurn: 1) and
// draws the discarded count.
func TestGenerateRielleTheEverwise(t *testing.T) {
	t.Parallel()
	power, toughness := "0", "3"
	card := &ScryfallCard{
		Name:      "Rielle, the Everwise",
		Layout:    "normal",
		ManaCost:  "{1}{U}{R}",
		TypeLine:  "Legendary Creature — Human Wizard",
		Power:     &power,
		Toughness: &toughness,
		OracleText: "Rielle gets +1/+0 for each instant and sorcery card in your graveyard.\n" +
			"Whenever you discard one or more cards for the first time each turn, draw that many cards.",
	}
	generatedSourceContains(t, card, []string{
		"game.EventCardDiscarded",
		"PlayerEventOrdinalThisTurn: 1,",
		"OneOrMore:                  true,",
		"Primitive: game.Draw{",
		"game.DynamicAmountEventCardCount",
		"game.DynamicAmountCountCardsInZone",
		"RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}",
	})
}

// Valiant Rescuer: a first-cycle-each-turn trigger ("Whenever you cycle another
// card for the first time each turn") that creates a token, plus Cycling. The
// cycle trigger gates on the first cycle occurrence of the turn.
func TestGenerateValiantRescuer(t *testing.T) {
	t.Parallel()
	power, toughness := "3", "1"
	card := &ScryfallCard{
		Name:      "Valiant Rescuer",
		Layout:    "normal",
		ManaCost:  "{1}{W}",
		TypeLine:  "Creature — Human Soldier",
		Power:     &power,
		Toughness: &toughness,
		OracleText: "Whenever you cycle another card for the first time each turn, create a 1/1 white Human Soldier creature token.\n" +
			"Cycling {2} ({2}, Discard this card: Draw a card.)",
	}
	generatedSourceContains(t, card, []string{
		"game.EventCycled",
		"PlayerEventOrdinalThisTurn: 1,",
		"ExcludeSelf:                true,",
		"game.CreateToken{",
	})
}

// card-type filter ("one or more nonland cards") that creates a Junk token.
func TestGenerateVeronicaDissidentScribe(t *testing.T) {
	t.Parallel()
	power, toughness := "3", "3"
	card := &ScryfallCard{
		Name:      "Veronica, Dissident Scribe",
		Layout:    "normal",
		ManaCost:  "{2}{R}",
		TypeLine:  "Legendary Creature — Human Artificer Rogue",
		Power:     &power,
		Toughness: &toughness,
		OracleText: "Whenever you discard one or more nonland cards for the first time each turn, create a Junk token. " +
			"(It's an artifact with \"{T}, Sacrifice this token: Exile the top card of your library. You may play that card this turn. Activate only as a sorcery.\")",
	}
	generatedSourceContains(t, card, []string{
		"game.EventCardDiscarded",
		"PlayerEventOrdinalThisTurn: 1,",
		"OneOrMore:                  true,",
	})
}
