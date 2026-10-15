package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WallOfDenial is the card definition for Wall of Denial.
//
// Type: Creature — Wall
// Cost: {1}{W}{U}
//
// Oracle text:
//
//	Defender, flying
//	Shroud (This creature can't be the target of spells or abilities.)
var WallOfDenial = newWallOfDenial

func newWallOfDenial() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Wall of Denial",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 8}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
				game.FlyingStaticBody,
				game.ShroudStaticBody,
			},
			OracleText: `
			Defender, flying
			Shroud (This creature can't be the target of spells or abilities.)
		`,
		},
	}
}
