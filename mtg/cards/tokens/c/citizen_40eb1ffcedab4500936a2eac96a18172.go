package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Citizen is the card definition for Citizen.
//
// Type: Token Creature — Citizen
//
// Oracle text:
//   This creature is all colors.

// CitizenToken40eb1ffcedab4500936a2eac96a18172 is the card definition for Citizen.
var CitizenToken40eb1ffcedab4500936a2eac96a18172 = newCitizenToken40eb1ffcedab4500936a2eac96a18172()

func newCitizenToken40eb1ffcedab4500936a2eac96a18172() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name:      "Citizen",
			Colors:    []color.Color{color.Black, color.Green, color.Red, color.Blue, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Citizen},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature is all colors.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerColor,
							SetColors:      []color.Color{color.White, color.Blue, color.Black, color.Red, color.Green},
						},
					},
				},
			},
			OracleText: `
			This creature is all colors.
		`,
		},
	}
}
