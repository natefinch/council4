package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Dwarf Token
//
// Type: Token Creature — Dwarf
//
// Oracle text:

// DwarfTokenTokenc92fc51f5d4e4e41bcbc4259d36f9f39 is the card definition for Dwarf Token.
var DwarfTokenTokenc92fc51f5d4e4e41bcbc4259d36f9f39 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Dwarf Token",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dwarf},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
