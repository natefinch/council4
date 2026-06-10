package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Zombie Mutant
//
// Type: Token Creature — Zombie Mutant
//
// Oracle text:

// ZombieMutantTokendd903bffd96f431a8ff988d0ad92f937 is the card definition for Zombie Mutant.
var ZombieMutantTokendd903bffd96f431a8ff988d0ad92f937 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Zombie Mutant",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Zombie, types.Mutant},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
