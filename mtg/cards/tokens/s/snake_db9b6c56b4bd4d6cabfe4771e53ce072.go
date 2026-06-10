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

// SnakeTokendb9b6c56b4bd4d6cabfe4771e53ce072 is the card definition for Snake.
var SnakeTokendb9b6c56b4bd4d6cabfe4771e53ce072 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Snake",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Snake},
		Power:     opt.Val(game.PT{Value: 5}),
		Toughness: opt.Val(game.PT{Value: 4}),
	},
}
