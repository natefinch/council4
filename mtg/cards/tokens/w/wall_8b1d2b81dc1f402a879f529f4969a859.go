package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Wall
//
// Type: Token Creature — Wall
//
// Oracle text:
//   Defender

// WallToken8b1d2b81dc1f402a879f529f4969a859 is the card definition for Wall.
var WallToken8b1d2b81dc1f402a879f529f4969a859 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Wall",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Wall},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{
			game.DefenderStaticBody,
		},
		OracleText: `
			Defender
		`,
	},
}
