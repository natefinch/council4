package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShrikeForce is the card definition for Shrike Force.
//
// Type: Creature — Bird Knight
// Cost: {2}{W}
//
// Oracle text:
//
//	Flying, double strike, vigilance
var ShrikeForce = newShrikeForce

func newShrikeForce() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Shrike Force",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird, types.Knight},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.DoubleStrikeStaticBody,
				game.VigilanceStaticBody,
			},
			OracleText: `
			Flying, double strike, vigilance
		`,
		},
	}
}
