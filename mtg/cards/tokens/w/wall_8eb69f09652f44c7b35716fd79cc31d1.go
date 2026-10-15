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
//   Defender, reach

// WallToken8eb69f09652f44c7b35716fd79cc31d1 is the card definition for Wall.
var WallToken8eb69f09652f44c7b35716fd79cc31d1 = newWallToken8eb69f09652f44c7b35716fd79cc31d1()

func newWallToken8eb69f09652f44c7b35716fd79cc31d1() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Wall",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
				game.ReachStaticBody,
			},
			OracleText: `
			Defender, reach
		`,
		},
	}
}
