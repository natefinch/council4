package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Primo, the Indivisible
//
// Type: Token Legendary Creature — Fractal
//
// Oracle text:

// PrimoTheIndivisibleToken19d20151ecb44627aa803086ad77b4a5 is the card definition for Primo, the Indivisible.
var PrimoTheIndivisibleToken19d20151ecb44627aa803086ad77b4a5 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue, color.Green),
	CardFace: game.CardFace{
		Name:       "Primo, the Indivisible",
		Colors:     []color.Color{color.Green, color.Blue},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Fractal},
		Power:      opt.Val(game.PT{Value: 0}),
		Toughness:  opt.Val(game.PT{Value: 0}),
	},
}
