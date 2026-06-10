package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Cordyceps Infected
//
// Type: Token Creature — Fungus Zombie
//
// Oracle text:

// CordycepsInfectedTokenf9feeb60d20a46e692e231b993719574 is the card definition for Cordyceps Infected.
var CordycepsInfectedTokenf9feeb60d20a46e692e231b993719574 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Cordyceps Infected",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Fungus, types.Zombie},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
