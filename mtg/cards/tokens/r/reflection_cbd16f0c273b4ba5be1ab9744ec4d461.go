package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Reflection
//
// Type: Token Creature — Reflection
//
// Oracle text:

// ReflectionTokencbd16f0c273b4ba5be1ab9744ec4d461 is the card definition for Reflection.
var ReflectionTokencbd16f0c273b4ba5be1ab9744ec4d461 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Reflection",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Reflection},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
