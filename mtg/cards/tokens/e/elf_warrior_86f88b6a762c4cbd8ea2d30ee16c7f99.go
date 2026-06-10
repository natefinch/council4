package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Elf Warrior
//
// Type: Token Creature — Elf Warrior
//
// Oracle text:

// ElfWarriorToken86f88b6a762c4cbd8ea2d30ee16c7f99 is the card definition for Elf Warrior.
var ElfWarriorToken86f88b6a762c4cbd8ea2d30ee16c7f99 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Green),
	CardFace: game.CardFace{
		Name:      "Elf Warrior",
		Colors:    []color.Color{color.Green, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elf, types.Warrior},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
