package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SkyspearCavalry is the card definition for Skyspear Cavalry.
//
// Type: Creature — Human Soldier
// Cost: {3}{W}{W}
//
// Oracle text:
//
//	Flying
//	Double strike (This creature deals both first-strike and regular combat damage.)
var SkyspearCavalry = newSkyspearCavalry

func newSkyspearCavalry() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Skyspear Cavalry",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.DoubleStrikeStaticBody,
			},
			OracleText: `
			Flying
			Double strike (This creature deals both first-strike and regular combat damage.)
		`,
		},
	}
}
