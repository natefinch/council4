package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Storm Crow
//
// Type: Token Creature — Bird
//
// Oracle text:
//   Flying

// StormCrowToken38d87b870c6747a58093b49aa11f6196 is the card definition for Storm Crow.
var StormCrowToken38d87b870c6747a58093b49aa11f6196 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Storm Crow",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Bird},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	},
}
