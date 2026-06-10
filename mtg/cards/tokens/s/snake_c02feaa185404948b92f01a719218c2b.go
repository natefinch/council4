package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Snake
//
// Type: Token Creature — Snake
//
// Oracle text:
//   Deathtouch

// SnakeTokenc02feaa185404948b92f01a719218c2b is the card definition for Snake.
var SnakeTokenc02feaa185404948b92f01a719218c2b = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Snake",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Snake},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.DeathtouchStaticBody,
		},
		OracleText: `
			Deathtouch
		`,
	},
}
