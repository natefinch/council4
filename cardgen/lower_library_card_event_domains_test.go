package cardgen

import "testing"

func TestLibraryCardPermanentOperationDoesNotRefuseCardDomainConsequences(t *testing.T) {
	assertCardPaths(t, &ScryfallCard{
		Name: "Observed Creature Consequence", Layout: "normal", TypeLine: "Enchantment",
		OracleText: "Whenever a creature attacks you, reveal the top card of your library. If it's a creature card, put it onto the battlefield tapped.",
	}, `Sequence[1].Primitive.(game.PutOnBattlefield).Source.linked = "sequence-effect-0-product"`)
}

func TestLibraryCardCombatConsumersRejectUnavailableOrCardOnlySubjects(t *testing.T) {
	for _, text := range []string{
		"Reveal the top card of your library. If it's a creature card, remove it from combat.",
		"Reveal the top card of your library. If it's a Forest card, remove that creature from combat.",
		"Whenever one or more creatures attack you, reveal the top card of your library. If it's a Forest card, remove that creature from combat.",
	} {
		t.Run(text, func(t *testing.T) {
			assertCardUnsupported(t, &ScryfallCard{
				Name: "Unavailable Combat Subject", Layout: "normal", TypeLine: "Enchantment", OracleText: text,
			})
		})
	}
}
