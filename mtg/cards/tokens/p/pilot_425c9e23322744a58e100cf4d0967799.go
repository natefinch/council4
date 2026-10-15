package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Pilot is the card definition for Pilot.
//
// Type: Token Creature — Pilot
//
// Oracle text:
//   This creature crews Vehicles as though its power were 2 greater.

// PilotToken425c9e23322744a58e100cf4d0967799 is the card definition for Pilot.
var PilotToken425c9e23322744a58e100cf4d0967799 = newPilotToken425c9e23322744a58e100cf4d0967799()

func newPilotToken425c9e23322744a58e100cf4d0967799() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Pilot",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Pilot},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text:           "This creature crews Vehicles as though its power were 2 greater.",
					CrewPowerBonus: 2,
				},
			},
			OracleText: `
			This creature crews Vehicles as though its power were 2 greater.
		`,
		},
	}
}
