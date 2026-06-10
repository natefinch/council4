package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Spider
//
// Type: Token Creature — Spider
//
// Oracle text:
//   Reach

// SpiderToken0c60a894a72446538f94ff150466b1de is the card definition for Spider.
var SpiderToken0c60a894a72446538f94ff150466b1de = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Spider",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Spider},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.ReachStaticBody,
		},
		OracleText: `
			Reach
		`,
	},
}
