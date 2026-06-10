package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Soldier
//
// Type: Token Creature — Soldier
//
// Oracle text:
//   Haste

// SoldierToken4f9f549126d34223b7f73e38048e4f09 is the card definition for Soldier.
var SoldierToken4f9f549126d34223b7f73e38048e4f09 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Red),
	CardFace: game.CardFace{
		Name:      "Soldier",
		Colors:    []color.Color{color.Red, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Soldier},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.HasteStaticBody,
		},
		OracleText: `
			Haste
		`,
	},
}
