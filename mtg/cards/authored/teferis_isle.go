package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

var TeferisIsle = cardset.Entry{
	OracleID:     "ce55657d-d82f-4528-a83e-5cad7de111fd",
	Name:         "Teferi's Isle",
	New:          newTeferisIsle,
	Verification: "mtg/rules/teferis_isle_test.go: TestTeferisIsleEntersTappedPhasesAndAddsTwoBlue",
}

func newTeferisIsle() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Teferi's Isle",
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Land},
			OracleText: "Phasing (This phases in or out before you untap during each of your untap steps. While it's phased out, it's treated as though it doesn't exist.)\n" +
				"Teferi's Isle enters tapped.\n" +
				"{T}: Add {U}{U}.",
			StaticAbilities: []game.StaticAbility{
				game.PhasingStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("Teferi's Isle enters tapped."),
			},
			ManaAbilities: []game.ManaAbility{{
				Text:            "{T}: Add {U}{U}.",
				AdditionalCosts: cost.Tap,
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.AddMana{
						Amount:    game.Fixed(2),
						ManaColor: mana.U,
					},
				}}}.Ability(),
			}},
		},
	}
}
