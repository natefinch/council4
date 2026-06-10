package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Satyr
//
// Type: Token Creature — Satyr
//
// Oracle text:
//   Haste

// SatyrToken0e23deae45194acb85b8538979f7eedc is the card definition for Satyr.
var SatyrToken0e23deae45194acb85b8538979f7eedc = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red, color.Green),
	CardFace: game.CardFace{
		Name:      "Satyr",
		Colors:    []color.Color{color.Green, color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Satyr},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.HasteStaticBody,
		},
		OracleText: `
			Haste
		`,
	},
}
