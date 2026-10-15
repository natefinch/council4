package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Rat is the card definition for Rat.
//
// Type: Token Creature — Rat
//
// Oracle text:
//   This token can't block.

// RatTokenb258066788dc4916a64c3080956b24bc is the card definition for Rat.
var RatTokenb258066788dc4916a64c3080956b24bc = newRatTokenb258066788dc4916a64c3080956b24bc()

func newRatTokenb258066788dc4916a64c3080956b24bc() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:      "Rat",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Rat},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.CantBlockStaticBody,
			},
			OracleText: `
			This token can't block.
		`,
		},
	}
}
