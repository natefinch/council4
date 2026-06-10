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

// SnakeTokenb54d7c256de9478293361bf27dc8c046 is the card definition for Snake.
var SnakeTokenb54d7c256de9478293361bf27dc8c046 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue, color.Green),
	CardFace: game.CardFace{
		Name:      "Snake",
		Colors:    []color.Color{color.Green, color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Snake},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
