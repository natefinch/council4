package cardgen

import "testing"

// Mary Read and Anne Bonny: a discard trigger whose card filter is a union of
// subtypes spanning multiple card types ("Whenever you discard an Island,
// Pirate, or Vehicle card, create a tapped Treasure token"). The union lowers
// to a SubtypesAny selection on the discarded card, and the payoff makes a
// tapped Treasure token.
func TestGenerateMaryReadAndAnneBonny(t *testing.T) {
	t.Parallel()
	power, toughness := "3", "3"
	card := &ScryfallCard{
		Name:      "Mary Read and Anne Bonny",
		Layout:    "normal",
		ManaCost:  "{1}{U}{R}",
		TypeLine:  "Legendary Creature — Human Assassin Pirate",
		Power:     &power,
		Toughness: &toughness,
		OracleText: "Haste\n" +
			"{T}: Draw a card, then discard a card.\n" +
			"Whenever you discard an Island, Pirate, or Vehicle card, create a tapped Treasure token.",
	}
	generatedSourceContains(t, card, []string{
		"game.EventCardDiscarded",
		"SubtypesAny: []types.Sub{types.Sub(\"Island\"), types.Sub(\"Pirate\"), types.Sub(\"Vehicle\")}",
		"game.CreateToken{",
		"EntryTapped: true,",
	})
}
