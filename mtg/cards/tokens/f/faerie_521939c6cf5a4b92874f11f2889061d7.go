package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Faerie
//
// Type: Token Creature — Faerie
//
// Oracle text:
//   Flying

// FaerieToken521939c6cf5a4b92874f11f2889061d7 is the card definition for Faerie.
var FaerieToken521939c6cf5a4b92874f11f2889061d7 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Faerie",
		Colors:    []color.Color{color.Black, color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Faerie},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	},
}
