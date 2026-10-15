package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Dwarf is the card definition for Dwarf.
//
// Type: Token Creature — Dwarf
//
// Oracle text:

// DwarfTokenc92fc51f5d4e4e41bcbc4259d36f9f39 is the card definition for Dwarf.
var DwarfTokenc92fc51f5d4e4e41bcbc4259d36f9f39 = newDwarfTokenc92fc51f5d4e4e41bcbc4259d36f9f39()

func newDwarfTokenc92fc51f5d4e4e41bcbc4259d36f9f39() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:      "Dwarf",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dwarf},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
