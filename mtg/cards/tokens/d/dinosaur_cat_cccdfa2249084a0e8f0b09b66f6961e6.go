package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Dinosaur Cat
//
// Type: Token Creature — Dinosaur Cat
//
// Oracle text:

// DinosaurCatTokencccdfa2249084a0e8f0b09b66f6961e6 is the card definition for Dinosaur Cat.
var DinosaurCatTokencccdfa2249084a0e8f0b09b66f6961e6 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Red),
	CardFace: game.CardFace{
		Name:      "Dinosaur Cat",
		Colors:    []color.Color{color.Red, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dinosaur, types.Cat},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
