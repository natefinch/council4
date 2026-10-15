package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WallOfSwords is the card definition for Wall of Swords.
//
// Type: Creature — Wall
// Cost: {3}{W}
//
// Oracle text:
//
//	Defender (This creature can't attack.)
//	Flying
var WallOfSwords = newWallOfSwords

func newWallOfSwords() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Wall of Swords",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
				game.FlyingStaticBody,
			},
			OracleText: `
			Defender (This creature can't attack.)
			Flying
		`,
		},
	}
}
