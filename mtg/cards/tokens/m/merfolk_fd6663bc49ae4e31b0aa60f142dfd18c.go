package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Merfolk
//
// Type: Token Creature — Merfolk
//
// Oracle text:
//   Hexproof (This creature can't be the target of spells or abilities your opponents control.)

// MerfolkTokenfd6663bc49ae4e31b0aa60f142dfd18c is the card definition for Merfolk.
var MerfolkTokenfd6663bc49ae4e31b0aa60f142dfd18c = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Merfolk",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Merfolk},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.HexproofStaticBody,
		},
		OracleText: `
			Hexproof (This creature can't be the target of spells or abilities your opponents control.)
		`,
	},
}
