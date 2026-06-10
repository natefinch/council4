package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Cleric
//
// Type: Token Enchantment Creature — Cleric
//
// Oracle text:

// ClericToken196f944465ac40728b9e233e8df98d08 is the card definition for Cleric.
var ClericToken196f944465ac40728b9e233e8df98d08 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Cleric",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Enchantment, types.Creature},
		Subtypes:  []types.Sub{types.Cleric},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
