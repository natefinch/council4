package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Alien
//
// Type: Token Creature — Alien
//
// Oracle text:

// AlienToken743d33fa03184f76a989af54434f2873 is the card definition for Alien.
var AlienToken743d33fa03184f76a989af54434f2873 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Alien",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Alien},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
