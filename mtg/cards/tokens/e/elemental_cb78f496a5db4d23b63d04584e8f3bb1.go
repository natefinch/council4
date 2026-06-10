package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Elemental
//
// Type: Token Creature — Elemental
//
// Oracle text:

// ElementalTokencb78f496a5db4d23b63d04584e8f3bb1 is the card definition for Elemental.
var ElementalTokencb78f496a5db4d23b63d04584e8f3bb1 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Elemental",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elemental},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 0}),
	},
}
