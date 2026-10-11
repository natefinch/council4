package cardgen

import (
	"fmt"
	"testing"
)

func TestLowerLibraryCardCompositionAcrossShells(t *testing.T) {
	t.Parallel()
	for _, producer := range []string{"Look at", "Reveal"} {
		for _, consumer := range []struct {
			text string
			path string
		}{
			{"draw a card", "Primitive.(game.Draw).Amount.fixed = 1"},
			{"you gain 2 life", "Primitive.(game.GainLife).Amount.fixed = 2"},
			{"put it into your hand", "Primitive.(game.MoveCard).Destination = zone.Hand"},
			{"put it on the bottom of your library", "Primitive.(game.MoveCard).DestinationBottom = true"},
			{"put it onto the battlefield", "Primitive.(game.PutOnBattlefield).Recipient.Val.kind = game.PlayerReferenceController"},
			{"put it onto the battlefield tapped under your control", "Primitive.(game.PutOnBattlefield).EntryTapped = true"},
		} {
			body := fmt.Sprintf("%s the top card of your library. Scry 1. If it's a land card, %s.", producer, consumer.text)
			for _, shell := range []struct {
				name, typeLine, prefix, suffix, path string
			}{
				{"spell", "Sorcery", "", "", "SpellAbility.Val"},
				{"activated", "Artifact", "{1}, {T}: ", "", "ActivatedAbilities[0].Content"},
				{"triggered", "Enchantment", "At the beginning of your upkeep, ", "", "TriggeredAbilities[0].Content"},
				{"mode", "Sorcery", "Choose one \u2014\n\u2022 ", "\n\u2022 Draw a card.", "SpellAbility.Val"},
			} {
				t.Run(producer+"/"+consumer.text+"/"+shell.name, func(t *testing.T) {
					t.Parallel()
					card := &ScryfallCard{
						Name: "Library Composition", Layout: "normal", TypeLine: shell.typeLine,
						OracleText: shell.prefix + body + shell.suffix,
					}
					primitive := "game.LookAtLibraryTop"
					if producer == "Reveal" {
						primitive = "game.Reveal"
					}
					path := shell.path + ".Modes[0].Sequence"
					assertCardPaths(t, card,
						path+"[0].Primitive.("+primitive+").PublishLinked",
						path+"[1].Primitive.(game.Scry).Amount.fixed = 1",
						path+"[2]."+consumer.path,
						path+"[2].Condition",
					)
				})
			}
		}
	}
}

func TestLowerLibraryCardConditionCompositions(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Reveal the top card of your library. If it's a snow card, draw a card.",
		"Look at the top card of your library. If it's a legendary card, draw a card.",
		"Reveal the top card of your library. If it's a colorless card, draw a card.",
		"Reveal the top card of your library. If it's a Forest card, draw a card.",
		"Reveal the top card of your library. If it's a Zombie card, draw a card.",
		"Reveal the top card of your library. If it's a Goblin permanent card, draw a card.",
		"Reveal the top card of your library. If it's snow, draw a card.",
		"Reveal the top card of your library. If it's a blue creature card, draw a card.",
		"Reveal the top card of your library. If it's a creature or land card, draw a card.",
		"Reveal the top card of your library. If it's a noncreature, nonland card, draw a card.",
		"Look at the top card of target opponent's library. If it's a nonland card, draw a card.",
		"You may reveal the top card of your library. If it's a creature card, draw a card.",
		"Reveal the top card of your library. If it's a land card, draw a card. Otherwise, you gain 2 life.",
		"Reveal the top card of your library. If it's a creature card, create a 1/1 green Saproling creature token. If it's a land card, put it onto the battlefield under your control.",
		"Reveal the top card of your library. Scry 1. If it's a land card, put it into your hand. Look at the top card of your library. Draw a card. If it's a creature card, you gain 2 life.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			assertCardPaths(t, &ScryfallCard{
				Name: "Typed Observation Composition", Layout: "normal", TypeLine: "Sorcery", OracleText: text,
			}, "SpellAbility.Val.Modes[0].Sequence[0].Primitive")
		})
	}
}

func TestLowerLibraryCardProducersAcrossShells(t *testing.T) {
	t.Parallel()
	for _, producer := range []string{"Look at", "Reveal", "You may look at", "You may reveal"} {
		for _, shell := range []struct {
			name, typeLine, prefix, suffix, path string
		}{
			{"spell", "Sorcery", "", "", "SpellAbility.Val"},
			{"activated", "Artifact", "{1}, {T}: ", "", "ActivatedAbilities[0].Content"},
			{"triggered", "Enchantment", "At the beginning of your upkeep, ", "", "TriggeredAbilities[0].Content"},
			{"mode", "Sorcery", "Choose one \u2014\n\u2022 ", "\n\u2022 Draw a card.", "SpellAbility.Val"},
		} {
			t.Run(producer+"/"+shell.name, func(t *testing.T) {
				t.Parallel()
				card := &ScryfallCard{
					Name: "Single Observation", Layout: "normal", TypeLine: shell.typeLine,
					OracleText: shell.prefix + producer + " the top card of your library." + shell.suffix,
				}
				primitive := "game.LookAtLibraryTop"
				if producer == "Reveal" || producer == "You may reveal" {
					primitive = "game.Reveal"
				}
				path := shell.path + ".Modes[0].Sequence[0]"
				paths := []string{path + ".Primitive.(" + primitive + ").Player.kind = game.PlayerReferenceController"}
				if producer == "You may look at" || producer == "You may reveal" {
					optionalPath := path + ".Optional = true"
					if shell.name == "triggered" {
						optionalPath = "TriggeredAbilities[0].Optional = true"
					}
					paths = append(paths, optionalPath)
				}
				assertCardPaths(t, card, paths...)
			})
		}
	}
}

func TestLowerLibraryCardConditionsRequireModeledObservation(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Scry 1. If it's a snow card, draw a card.",
		"Scry 1. If it's a legendary card, draw a card.",
		"Scry 1. If it's a colorless card, draw a card.",
		"Look at the top card of your library. If it's a snow and legendary card, draw a card.",
		"Look at the top card of your library. If it's a colorless beautiful card, draw a card.",
		"Choose one \u2014\n\u2022 Reveal the top card of your library.\n\u2022 If it's a snow card, draw a card.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			assertCardUnsupported(t, &ScryfallCard{
				Name: "Observation Near Miss", Layout: "normal", TypeLine: "Sorcery", OracleText: text,
			})
		})
	}
}
