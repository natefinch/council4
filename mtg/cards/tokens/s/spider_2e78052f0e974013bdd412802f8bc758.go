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

// SpiderToken2e78052f0e974013bdd412802f8bc758 is the card definition for Spider.
var SpiderToken2e78052f0e974013bdd412802f8bc758 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black, color.Green),
	CardFace: game.CardFace{
		Name:      "Spider",
		Colors:    []color.Color{color.Black, color.Green},
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
