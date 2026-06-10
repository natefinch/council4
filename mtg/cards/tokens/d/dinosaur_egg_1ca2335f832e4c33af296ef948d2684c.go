package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Dinosaur Egg
//
// Type: Token Creature — Dinosaur Egg
//
// Oracle text:

// DinosaurEggToken1ca2335f832e4c33af296ef948d2684c is the card definition for Dinosaur Egg.
var DinosaurEggToken1ca2335f832e4c33af296ef948d2684c = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Dinosaur Egg",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dinosaur, types.Egg},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
