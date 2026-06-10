package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Elemental
//
// Type: Token Creature — Elemental
//
// Oracle text:

// ElementalTokenc5a790eba3cf4214b7439123ad002ade is the card definition for Elemental.
var ElementalTokenc5a790eba3cf4214b7439123ad002ade = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Elemental",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elemental},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
	},
}
