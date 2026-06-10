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
//   Menace, reach

// SpiderToken3e6a1c1316f04256b4681f537de78629 is the card definition for Spider.
var SpiderToken3e6a1c1316f04256b4681f537de78629 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Spider",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Spider},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.MenaceStaticBody,
			game.ReachStaticBody,
		},
		OracleText: `
			Menace, reach
		`,
	},
}
