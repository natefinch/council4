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
//   Defender, flying

// WallToken3a4be250dfef46c4b3f16e817e26eb97 is the card definition for Wall.
var WallToken3a4be250dfef46c4b3f16e817e26eb97 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Wall",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Wall},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{
			game.DefenderStaticBody,
			game.FlyingStaticBody,
		},
		OracleText: `
			Defender, flying
		`,
	},
}
