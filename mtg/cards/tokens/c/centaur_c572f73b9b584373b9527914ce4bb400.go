package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Centaur
//
// Type: Token Creature — Centaur
//
// Oracle text:
//   Protection from black

// CentaurTokenc572f73b9b584373b9527914ce4bb400 is the card definition for Centaur.
var CentaurTokenc572f73b9b584373b9527914ce4bb400 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Centaur",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Centaur},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{
			game.ProtectionFromColorsStaticAbility(color.Black),
		},
		OracleText: `
			Protection from black
		`,
	},
}
