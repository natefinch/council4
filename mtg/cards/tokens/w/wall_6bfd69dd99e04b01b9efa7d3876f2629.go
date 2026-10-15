package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Wall is the card definition for Wall.
//
// Type: Token Creature — Wall
//
// Oracle text:
//   Defender

// WallToken6bfd69dd99e04b01b9efa7d3876f2629 is the card definition for Wall.
var WallToken6bfd69dd99e04b01b9efa7d3876f2629 = newWallToken6bfd69dd99e04b01b9efa7d3876f2629()

func newWallToken6bfd69dd99e04b01b9efa7d3876f2629() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Wall",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
			OracleText: `
			Defender
		`,
		},
	}
}
