package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Carnivore
//
// Type: Token Creature — Beast
//
// Oracle text:

// CarnivoreToken09bb0fec5d7148438aee4505d9c84c82 is the card definition for Carnivore.
var CarnivoreToken09bb0fec5d7148438aee4505d9c84c82 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Carnivore",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Beast},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
