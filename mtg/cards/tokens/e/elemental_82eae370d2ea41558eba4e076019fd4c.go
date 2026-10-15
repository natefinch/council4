package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Elemental is the card definition for Elemental.
//
// Type: Token Creature — Elemental
//
// Oracle text:
//   This token is all colors.

// ElementalToken82eae370d2ea41558eba4e076019fd4c is the card definition for Elemental.
var ElementalToken82eae370d2ea41558eba4e076019fd4c = newElementalToken82eae370d2ea41558eba4e076019fd4c()

func newElementalToken82eae370d2ea41558eba4e076019fd4c() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name:      "Elemental",
			Colors:    []color.Color{color.Black, color.Green, color.Red, color.Blue, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This token is all colors.",
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
			This token is all colors.
		`,
		},
	}
}
