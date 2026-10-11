package cardgen

import "testing"

func TestChosenNameFreshFramePublicPrefixes(t *testing.T) {
	const choice = "Choose a land you control. Search your library for up to two basic land cards which have the same name as the chosen land, put them onto the battlefield tapped, then shuffle."
	for _, prefix := range []string{
		"Manifest the top card of your library.",
		"Bolster 1.",
		"Choose a creature type.",
	} {
		t.Run(prefix, func(t *testing.T) {
			assertCardPaths(t, &ScryfallCard{
				Name: "Fresh name choice", Layout: "normal", TypeLine: "Sorcery", OracleText: prefix + " " + choice,
			},
				"Primitive.(game.Choose).Choice.Kind = game.ResolutionChoicePermanent",
				"Primitive.(game.Choose).PublishChoice = game.ResolutionChosenPermanentChoiceKey",
				"Primitive.(game.Search).Spec.NameFromChoice = game.ResolutionChosenPermanentChoiceKey",
			)
		})
	}
}

func TestChosenNameFreshFramePublicRefusals(t *testing.T) {
	const search = "Search your library for up to two basic land cards which have the same name as the chosen land, put them onto the battlefield tapped, then shuffle."
	for _, text := range []string{
		"Choose a land you control. Bolster 1. " + search,
		"Choose a land you control. Manifest the top card of your library. " + search,
		"Choose a land you control. Choose a creature type. " + search,
		"Bolster 1. Choose a land you control. Choose a land you control. " + search,
		"Bolster 1. Choose a land you control. " + search + " Choose a land you control.",
		"Bolster 1. If you control a creature, choose a land you control. " + search,
		"Perform an unmodeled action. Choose a land you control. " + search,
	} {
		t.Run(text, func(t *testing.T) {
			card := &ScryfallCard{Name: "Fresh frame refusal", Layout: "normal", TypeLine: "Sorcery", OracleText: text}
			assertCardUnsupported(t, card)
			defs, diagnostics, err := CompileCardDefs(card)
			if err != nil || len(diagnostics) == 0 || len(defs) != 0 {
				t.Fatal("unmodeled prefix/frame acquired a partial executable CardDef")
			}
		})
	}
}
