package cardgen

import (
	"fmt"
	"testing"
)

func TestLowerLibraryCardExplicitObservationOwnership(t *testing.T) {
	for _, tc := range []struct {
		name, text, publisher string
		move, producer        int
	}{
		{"look before reveal", "Look at the top card of your library. Reveal the top card of target opponent's library. Put the looked-at card into your hand.", "game.LookAtLibraryTop", 2, 0},
		{"reveal before look", "Reveal the top card of your library. Look at the top card of target opponent's library. Put the revealed card into your hand.", "game.Reveal", 2, 0},
		{"looked card and its owner", "Look at the top card of target player's library. Reveal the top card of target opponent's library. Put the looked-at card into that player's hand.", "game.LookAtLibraryTop", 2, 0},
		{"revealed card and its owner", "Reveal the top card of target player's library. Look at the top card of target opponent's library. Put the revealed card into that player's hand.", "game.Reveal", 2, 0},
		{"generic card remains latest", "Look at the top card of your library. Reveal the top card of target opponent's library. Put that card into your hand.", "game.LookAtLibraryTop", 2, 1},
		{"linked reveal preserves card", "Look at the top card of your library. Reveal it. Put the revealed card into your hand.", "game.LookAtLibraryTop", 2, 0},
		{"linked reveal of earlier look", "Look at the top card of your library. Reveal the top card of target opponent's library. Reveal the looked-at card. Put the revealed card into your hand.", "game.LookAtLibraryTop", 3, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			card := &ScryfallCard{Name: "Explicit Observation Ownership", Layout: "normal", TypeLine: "Sorcery", OracleText: tc.text}
			assertCardPaths(t, card,
				"Sequence[0].Primitive.("+tc.publisher+").PublishLinked = \"sequence-effect-0-product\"",
				fmt.Sprintf("Sequence[%d].Primitive.(game.MoveCard).Card.LinkID = \"sequence-effect-%d-product\"", tc.move, tc.producer),
				fmt.Sprintf("Sequence[%d].Primitive.(game.MoveCard).Destination = zone.Hand", tc.move),
			)
		})
	}
	for _, text := range []string{
		"Look at the top card of your library. Put the revealed card into your hand.",
		"Reveal the top card of your library. Put the looked-at card into your hand.",
		"Look at the top card of your library. Reveal the top two cards of target opponent's library. Put the looked-at card into your hand.",
		"Look at the top card of your library. Exile target creature. Put the looked-at card into your hand.",
	} {
		t.Run("unsupported/"+text, func(t *testing.T) {
			assertCardUnsupported(t, &ScryfallCard{
				Name: "Explicit Observation Near Miss", Layout: "normal", TypeLine: "Sorcery", OracleText: text,
			})
		})
	}
	t.Run("independent products keep owners", func(t *testing.T) {
		assertCardPaths(t, &ScryfallCard{
			Name: "Independent Observation Owners", Layout: "normal", TypeLine: "Sorcery",
			OracleText: "Look at the top card of target player's library. Reveal the top card of target opponent's library. Put the looked-at card into that player's hand. Put the revealed card into that player's graveyard.",
		},
			"Sequence[0].Primitive.(game.LookAtLibraryTop).PublishLinked = \"sequence-effect-0-product\"",
			"Sequence[1].Primitive.(game.Reveal).PublishLinked = \"sequence-effect-1-product\"",
			"Sequence[2].Primitive.(game.MoveCard).Card.LinkID = \"sequence-effect-0-product\"",
			"Sequence[2].Primitive.(game.MoveCard).Destination = zone.Hand",
			"Sequence[3].Primitive.(game.MoveCard).Card.LinkID = \"sequence-effect-1-product\"",
			"Sequence[3].Primitive.(game.MoveCard).Destination = zone.Graveyard",
		)
	})
}

func TestLowerSingletonLifePaymentTriggerChoice(t *testing.T) {
	payment := &ScryfallCard{
		Name: "Singleton Trigger Payment", Layout: "normal", TypeLine: "Enchantment",
		OracleText: "At the beginning of your upkeep, you may pay 2 life.",
	}
	assertCardPaths(t, payment,
		"TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.Pay).Payment.Payer.Val.kind = game.PlayerReferenceController",
		"TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.Pay).Payment.AdditionalCosts[0].Kind = cost.AdditionalPayLife",
		"TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.Pay).Payment.AdditionalCosts[0].Amount = 2",
	)
	assertCardPathsAbsent(t, payment, "TriggeredAbilities[0].Optional = true", "Sequence[0].Optional = true")
	assertCardPaths(t, &ScryfallCard{
		Name: "Ordinary Optional Trigger", Layout: "normal", TypeLine: "Enchantment",
		OracleText: "At the beginning of your upkeep, you may draw a card.",
	}, "TriggeredAbilities[0].Optional = true", "Sequence[0].Primitive.(game.Draw)")
	rider := &ScryfallCard{
		Name: "Independent Payment Rider", Layout: "normal", TypeLine: "Enchantment",
		OracleText: "At the beginning of your upkeep, you may pay 2 life. You may draw a card.",
	}
	assertCardPaths(t, rider, "Sequence[0].Primitive.(game.Pay)", "Sequence[1].Optional = true", "Sequence[1].Primitive.(game.Draw)")
	assertCardPathsAbsent(t, rider, "TriggeredAbilities[0].Optional = true", "Sequence[0].Optional = true")
}
