package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Ooze
//
// Type: Token Creature — Ooze
//
// Oracle text:
//   Trample

// OozeToken6d0b29d44d4b4307977827c7483cbcb8 is the card definition for Ooze.
var OozeToken6d0b29d44d4b4307977827c7483cbcb8 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Ooze",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Ooze},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 0}),
		StaticAbilities: []game.StaticAbility{
			game.TrampleStaticBody,
		},
		OracleText: `
			Trample
		`,
	},
}
