package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Rabbit
//
// Type: Token Creature — Rabbit
//
// Oracle text:

// RabbitTokenfee35fc4a5b54c2c8c6a62dad166ea72 is the card definition for Rabbit.
var RabbitTokenfee35fc4a5b54c2c8c6a62dad166ea72 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Rabbit",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Rabbit},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
