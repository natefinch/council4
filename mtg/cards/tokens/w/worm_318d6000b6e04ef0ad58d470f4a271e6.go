package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Worm
//
// Type: Token Creature — Worm
//
// Oracle text:

// WormToken318d6000b6e04ef0ad58d470f4a271e6 is the card definition for Worm.
var WormToken318d6000b6e04ef0ad58d470f4a271e6 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black, color.Green),
	CardFace: game.CardFace{
		Name:      "Worm",
		Colors:    []color.Color{color.Black, color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Worm},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
