package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Rhino
//
// Type: Token Creature — Rhino
//
// Oracle text:
//   Trample

// RhinoTokena685171a616f42e191614a64fb51a977 is the card definition for Rhino.
var RhinoTokena685171a616f42e191614a64fb51a977 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Rhino",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Rhino},
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
