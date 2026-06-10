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

// ElementalTokena9ea980e473445e5b5b74ec092db3299 is the card definition for Elemental.
var ElementalTokena9ea980e473445e5b5b74ec092db3299 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red, color.Green),
	CardFace: game.CardFace{
		Name:      "Elemental",
		Colors:    []color.Color{color.Green, color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elemental},
		Power:     opt.Val(game.PT{Value: 5}),
		Toughness: opt.Val(game.PT{Value: 5}),
	},
}
