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
//   Trample, haste

// ElementalToken6a5497df7ad8443bafc90017f9b1eac0 is the card definition for Elemental.
var ElementalToken6a5497df7ad8443bafc90017f9b1eac0 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Elemental",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elemental},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.TrampleStaticBody,
			game.HasteStaticBody,
		},
		OracleText: `
			Trample, haste
		`,
	},
}
