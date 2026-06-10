package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Fungus Beast
//
// Type: Token Creature — Fungus Beast
//
// Oracle text:
//   Trample

// FungusBeastToken81183cc6d39c4e51a9822351588987aa is the card definition for Fungus Beast.
var FungusBeastToken81183cc6d39c4e51a9822351588987aa = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Fungus Beast",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Fungus, types.Beast},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{
			game.TrampleStaticBody,
		},
		OracleText: `
			Trample
		`,
	},
}
