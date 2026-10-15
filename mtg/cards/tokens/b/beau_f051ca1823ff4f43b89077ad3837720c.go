package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Beau is the card definition for Beau.
//
// Type: Token Legendary Creature — Ox
//
// Oracle text:
//   This creature's power and toughness are each equal to the number of lands you control.

// BeauTokenf051ca1823ff4f43b89077ad3837720c is the card definition for Beau.
var BeauTokenf051ca1823ff4f43b89077ad3837720c = newBeauTokenf051ca1823ff4f43b89077ad3837720c()

func newBeauTokenf051ca1823ff4f43b89077ad3837720c() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:             "Beau",
			Colors:           []color.Color{color.Blue},
			Supertypes:       []types.Super{types.Legendary},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Ox},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerLandCount}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerLandCount}),
			OracleText: `
			This creature's power and toughness are each equal to the number of lands you control.
		`,
		},
	}
}
