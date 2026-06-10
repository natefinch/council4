package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Elf
//
// Type: Token Creature — Elf
//
// Oracle text:

// ElfTokenc52d1bba36d54eccbea2847285e0cb71 is the card definition for Elf.
var ElfTokenc52d1bba36d54eccbea2847285e0cb71 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Elf",
		Colors:    []color.Color{color.Black, color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elf},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
