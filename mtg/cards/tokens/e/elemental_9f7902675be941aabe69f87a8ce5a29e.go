package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Elemental is the card definition for Elemental.
//
// Type: Token Creature — Elemental
//
// Oracle text:
//   This creature's power and toughness are each equal to the number of creatures you control.

// ElementalToken9f7902675be941aabe69f87a8ce5a29e is the card definition for Elemental.
var ElementalToken9f7902675be941aabe69f87a8ce5a29e = newElementalToken9f7902675be941aabe69f87a8ce5a29e()

func newElementalToken9f7902675be941aabe69f87a8ce5a29e() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name:             "Elemental",
			Colors:           []color.Color{color.Green, color.White},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Elemental},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCount}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCount}),
			OracleText: `
			This creature's power and toughness are each equal to the number of creatures you control.
		`,
		},
	}
}
