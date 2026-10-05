package cardgen

import "testing"

func TestLibraryCardPaymentUsesSelectedPlayerObservation(t *testing.T) {
	assertCardPaths(t, &ScryfallCard{
		Name: "Selected Library Payment", Layout: "normal", TypeLine: "Artifact",
		OracleText: "{T}: Look at the top card of target player's library. If it's a nonland card, you may pay 2 life. If you do, put it into your graveyard.",
	},
		"ActivatedAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.LookAtLibraryTop).Player.kind = game.PlayerReferenceTargetPlayer",
		"ActivatedAbilities[0].Content.Modes[0].Sequence[1].Primitive.(game.Pay).Payment.AdditionalCosts[0].Kind = cost.AdditionalPayLife",
		"ActivatedAbilities[0].Content.Modes[0].Sequence[1].Primitive.(game.Pay).Payment.AdditionalCosts[0].Amount = 2",
		"ActivatedAbilities[0].Content.Modes[0].Sequence[2].Primitive.(game.MoveCard).Card.LinkID = \"sequence-effect-0-product\"",
		"ActivatedAbilities[0].Content.Modes[0].Sequence[2].Primitive.(game.MoveCard).Destination = zone.Graveyard",
	)
	assertCardPathsAbsent(t, &ScryfallCard{
		Name: "Selected Library Payment", Layout: "normal", TypeLine: "Artifact",
		OracleText: "{T}: Look at the top card of target player's library. If it's a nonland card, you may pay 2 life. If you do, put it into your graveyard.",
	}, "ActivatedAbilities[0].Content.Modes[0].Sequence[1].Optional = true")
}

func TestLibraryCardOptionalLossIsNotPayment(t *testing.T) {
	card := &ScryfallCard{
		Name: "Selected Library Loss", Layout: "normal", TypeLine: "Artifact",
		OracleText: "{T}: Look at the top card of target player's library. If it's a nonland card, you may lose 2 life. If you do, put it into your graveyard.",
	}
	assertCardPaths(t, card,
		"Sequence[1].Primitive.(game.LoseLife).Amount.fixed = 2",
		"Sequence[1].Optional = true",
	)
	assertCardPathsAbsent(t, card, "Primitive.(game.Pay)")
}

func TestLibraryCardPaymentRefusesUnmodeledCosts(t *testing.T) {
	for _, payment := range []string{"pay life equal to its power", "pay 2 life for each creature you control", "pay 2 life or discard a card"} {
		t.Run(payment, func(t *testing.T) {
			assertCardUnsupported(t, &ScryfallCard{
				Name: "Unmodeled Library Payment", Layout: "normal", TypeLine: "Artifact",
				OracleText: "{T}: Look at the top card of target player's library. If it's a nonland card, you may " + payment + ". If you do, put it into your graveyard.",
			})
		})
	}

}

func TestLibraryCardPaymentUsesTypedLibraryOwnerDestination(t *testing.T) {
	assertCardPaths(t, &ScryfallCard{
		Name: "Wand of Denial", Layout: "normal", TypeLine: "Artifact", ManaCost: "{2}",
		OracleText: "{T}: Look at the top card of target player's library. If it's a nonland card, you may pay 2 life. If you do, put it into that player's graveyard.",
	},
		"Sequence[0].Primitive.(game.LookAtLibraryTop).Player.kind = game.PlayerReferenceTargetPlayer",
		"Sequence[1].Primitive.(game.Pay).Payment.Payer.Val.kind = game.PlayerReferenceController",
		"Sequence[2].Primitive.(game.MoveCard).Card.LinkID = \"sequence-effect-0-product\"",
		"Sequence[2].Primitive.(game.MoveCard).Destination = zone.Graveyard",
	)
}
