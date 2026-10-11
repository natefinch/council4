package cardgen

import "testing"

func TestChosenNameUnmodeledChoiceGrammarRefusesPublic(t *testing.T) {
	const search = "Search your library for up to two basic land cards which have the same name as the chosen land, put them onto the battlefield tapped, then shuffle."
	for _, choice := range []string{
		"Choose a land at random from among lands you control. ",
		"Choose a land secretly from among lands you control. ",
	} {
		t.Run(choice, func(t *testing.T) {
			assertCardUnsupported(t, &ScryfallCard{
				Name: "Unmodeled chosen name", Layout: "normal", TypeLine: "Sorcery",
				OracleText: choice + "You gain 2 life. " + search,
			})
		})
	}
}
