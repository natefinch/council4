package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Elemental
//
// Type: Token Enchantment Creature — Elemental
//
// Oracle text:

// ElementalTokene43cada01ed74246ae53c34a0e91a012 is the card definition for Elemental.
var ElementalTokene43cada01ed74246ae53c34a0e91a012 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Elemental",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Enchantment, types.Creature},
		Subtypes:  []types.Sub{types.Elemental},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
