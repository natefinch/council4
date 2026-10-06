package cardgen

import "testing"

func TestLibraryCardOptionalProducerKeepsMandatoryConsumers(t *testing.T) {
	for _, text := range []string{
		"You may reveal the top card of your library. If it's a creature card, draw a card. You gain 2 life.",
		"You may look at the top card of your library. If it's a land card, put it into your hand. You gain 2 life.",
	} {
		t.Run(text, func(t *testing.T) {
			card := &ScryfallCard{
				Name: "Optional Observation", Layout: "normal", TypeLine: "Sorcery", OracleText: text,
			}
			assertCardPaths(t, card,
				"SpellAbility.Val.Modes[0].Sequence[0].Optional = true",
				"SpellAbility.Val.Modes[0].Sequence[2].Primitive.(game.GainLife).Amount.fixed = 2",
			)
			assertCardPathsAbsent(t, card,
				"SpellAbility.Val.Modes[0].Sequence[1].Optional",
				"SpellAbility.Val.Modes[0].Sequence[2].Optional",
				"SpellAbility.Val.Modes[0].Sequence[2].Condition",
				"SpellAbility.Val.Modes[0].Sequence[2].ResultGate",
			)
		})
	}
}

func TestLibraryCardOptionalRevealAndMoveIsOneChoice(t *testing.T) {
	for _, noun := range []string{"land card", "snow card", "Zombie card", "instant or sorcery card"} {
		t.Run(noun, func(t *testing.T) {
			card := &ScryfallCard{
				Name: "Optional Observed Card Action", Layout: "normal", TypeLine: "Artifact",
				OracleText: "{T}: Look at the top card of your library. If it's a " + noun + ", you may reveal it and put it into your hand.",
			}
			assertCardPaths(t, card,
				"ActivatedAbilities[0].Content.Modes[0].Sequence[1].Optional = true",
				"ActivatedAbilities[0].Content.Modes[0].Sequence[1].PublishOptionalDecision",
				"ActivatedAbilities[0].Content.Modes[0].Sequence[2].OptionalDecisionGate",
				"ActivatedAbilities[0].Content.Modes[0].Sequence[2].Primitive.(game.MoveCard).Destination = zone.Hand",
			)
			assertCardPathsAbsent(t, card, "ActivatedAbilities[0].Content.Modes[0].Sequence[2].Optional = true")
		})
	}
}
