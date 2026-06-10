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
//   Vigilance

// ElementalToken3edbadbb2efb43a9aad58d1019d1ab6e is the card definition for Elemental.
var ElementalToken3edbadbb2efb43a9aad58d1019d1ab6e = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Green),
	CardFace: game.CardFace{
		Name:      "Elemental",
		Colors:    []color.Color{color.Green, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elemental},
		Power:     opt.Val(game.PT{Value: 8}),
		Toughness: opt.Val(game.PT{Value: 8}),
		StaticAbilities: []game.StaticAbility{
			game.VigilanceStaticBody,
		},
		OracleText: `
			Vigilance
		`,
	},
}
