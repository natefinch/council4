package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Fractal
//
// Type: Token Creature — Fractal
//
// Oracle text:

// FractalToken23c6a041b8ff449c8a2ce3939668c017 is the card definition for Fractal.
var FractalToken23c6a041b8ff449c8a2ce3939668c017 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue, color.Green),
	CardFace: game.CardFace{
		Name:      "Fractal",
		Colors:    []color.Color{color.Green, color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Fractal},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 0}),
	},
}
