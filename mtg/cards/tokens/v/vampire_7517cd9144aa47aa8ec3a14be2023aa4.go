package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Vampire
//
// Type: Token Creature — Vampire
//
// Oracle text:
//   Lifelink

// VampireToken7517cd9144aa47aa8ec3a14be2023aa4 is the card definition for Vampire.
var VampireToken7517cd9144aa47aa8ec3a14be2023aa4 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Black),
	CardFace: game.CardFace{
		Name:      "Vampire",
		Colors:    []color.Color{color.Black, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Vampire},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.LifelinkStaticBody,
		},
		OracleText: `
			Lifelink
		`,
	},
}
