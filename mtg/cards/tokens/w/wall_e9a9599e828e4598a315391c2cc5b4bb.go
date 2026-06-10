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

// WallTokene9a9599e828e4598a315391c2cc5b4bb is the card definition for Wall.
var WallTokene9a9599e828e4598a315391c2cc5b4bb = &game.CardDef{
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
		},
		OracleText: `
			Defender
		`,
	},
}
