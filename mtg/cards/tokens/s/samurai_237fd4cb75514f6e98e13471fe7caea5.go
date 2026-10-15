package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Samurai is the card definition for Samurai.
//
// Type: Token Creature — Samurai
//
// Oracle text:
//   Double strike

// SamuraiToken237fd4cb75514f6e98e13471fe7caea5 is the card definition for Samurai.
var SamuraiToken237fd4cb75514f6e98e13471fe7caea5 = newSamuraiToken237fd4cb75514f6e98e13471fe7caea5()

func newSamuraiToken237fd4cb75514f6e98e13471fe7caea5() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:      "Samurai",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Samurai},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.DoubleStrikeStaticBody,
			},
			OracleText: `
			Double strike
		`,
		},
	}
}
