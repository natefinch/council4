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

// SpiderTokenc16b3a1e0e6041f2899ed22bcf1144ae is the card definition for Spider.
var SpiderTokenc16b3a1e0e6041f2899ed22bcf1144ae = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Spider",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Spider},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{
			game.ReachStaticBody,
		},
		OracleText: `
			Reach
		`,
	},
}
