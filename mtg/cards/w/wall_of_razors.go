package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WallOfRazors is the card definition for Wall of Razors.
//
// Type: Creature — Wall
// Cost: {1}{R}
//
// Oracle text:
//
//	Defender (This creature can't attack.)
//	First strike
var WallOfRazors = newWallOfRazors

func newWallOfRazors() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Wall of Razors",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
				game.FirstStrikeStaticBody,
			},
			OracleText: `
			Defender (This creature can't attack.)
			First strike
		`,
		},
	}
}
