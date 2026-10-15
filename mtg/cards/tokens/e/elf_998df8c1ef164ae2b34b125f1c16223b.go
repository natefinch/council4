package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Elf is the card definition for Elf.
//
// Type: Token Creature — Elf
//
// Oracle text:

// ElfToken998df8c1ef164ae2b34b125f1c16223b is the card definition for Elf.
var ElfToken998df8c1ef164ae2b34b125f1c16223b = newElfToken998df8c1ef164ae2b34b125f1c16223b()

func newElfToken998df8c1ef164ae2b34b125f1c16223b() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:      "Elf",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
