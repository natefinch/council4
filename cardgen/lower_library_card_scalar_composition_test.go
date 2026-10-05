package cardgen

import "testing"

func TestLibraryCardCharacteristicConsumersComposeAcrossAbilityShells(t *testing.T) {
	body := "Look at the top card of your library. If it's a creature card, you gain life equal to its toughness."
	for _, tc := range []struct {
		name string
		text string
		path string
	}{
		{"spell", body, "SpellAbility.Val.Modes[0]"},
		{"activated", "{T}: " + body, "ActivatedAbilities[0].Content.Modes[0]"},
		{"triggered", "At the beginning of your upkeep, " + body, "TriggeredAbilities[0].Content.Modes[0]"},
		{"modal", "Choose one \u2014\n\u2022 " + body + "\n\u2022 Draw a card.", "SpellAbility.Val.Modes[0]"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			card := &ScryfallCard{Name: "Characteristic Shell", Layout: "normal", TypeLine: "Artifact", OracleText: tc.text}
			if tc.name == "spell" || tc.name == "modal" {
				card.TypeLine = "Sorcery"
			}
			assertCardPaths(t, card,
				tc.path+`.Sequence[0].Primitive.(game.LookAtLibraryTop).PublishCharacteristics.Toughness = "sequence-effect-0-toughness"`,
				tc.path+`.Sequence[1].ResultGate.Val.Key = "sequence-effect-0-toughness"`,
				tc.path+".Sequence[1].ResultGate.Val.AmountAvailable = true",
			)
		})
	}
}

func TestLibraryCardCharacteristicsSupportPostDrawAndPreMoveGroups(t *testing.T) {
	power, toughness := "2", "5"
	for _, card := range []*ScryfallCard{
		{Name: "Nissa's Revelation", Layout: "normal", TypeLine: "Sorcery", ManaCost: "{5}{G}{G}",
			OracleText: "Scry 5, then reveal the top card of your library. If it's a creature card, you draw cards equal to its power and you gain life equal to its toughness."},
		{Name: "Sapling of Colfenor", Layout: "normal", TypeLine: "Legendary Creature - Treefolk Shaman", ManaCost: "{3}{B}{G}", Power: &power, Toughness: &toughness,
			OracleText: "Indestructible\nWhenever Sapling of Colfenor attacks, reveal the top card of your library. If it's a creature card, you gain life equal to that card's toughness, lose life equal to its power, then put it into your hand."},
	} {
		t.Run(card.Name, func(t *testing.T) {
			assertCardPaths(t, card, "PublishCharacteristics.Power", "PublishCharacteristics.Toughness",
				"Primitive.(game.GainLife)", "ResultGate.Val.AmountAvailable = true")
		})
	}
}

func TestLibraryCardCharacteristicConsumersRefuseUnmodeledOperands(t *testing.T) {
	for _, body := range []string{
		"you gain life equal to twice its power",
		"you gain life equal to its power plus its toughness",
		"you gain life equal to its power this turn",
		"you may pay life equal to its power",
	} {
		t.Run(body, func(t *testing.T) {
			assertCardUnsupported(t, &ScryfallCard{
				Name: "Unmodeled Characteristic Operand", Layout: "normal", TypeLine: "Sorcery",
				OracleText: "Reveal the top card of your library. If it's a creature card, " + body + ".",
			})
		})
	}
}
