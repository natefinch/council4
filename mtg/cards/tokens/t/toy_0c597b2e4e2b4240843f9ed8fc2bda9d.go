package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Toy
//
// Type: Token Artifact Creature — Toy
//
// Oracle text:

// ToyToken0c597b2e4e2b4240843f9ed8fc2bda9d is the card definition for Toy.
var ToyToken0c597b2e4e2b4240843f9ed8fc2bda9d = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Toy",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Toy},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
