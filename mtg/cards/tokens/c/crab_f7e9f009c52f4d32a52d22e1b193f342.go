package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Crab
//
// Type: Token Creature — Crab
//
// Oracle text:

// CrabTokenf7e9f009c52f4d32a52d22e1b193f342 is the card definition for Crab.
var CrabTokenf7e9f009c52f4d32a52d22e1b193f342 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Crab",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Crab},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 3}),
	},
}
