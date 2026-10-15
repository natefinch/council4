package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RipscalePredator is the card definition for Ripscale Predator.
//
// Type: Creature — Dinosaur
// Cost: {4}{R}{R}
//
// Oracle text:
//
//	Menace (This creature can't be blocked except by two or more creatures.)
var RipscalePredator = newRipscalePredator

func newRipscalePredator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Ripscale Predator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dinosaur},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
			},
			OracleText: `
			Menace (This creature can't be blocked except by two or more creatures.)
		`,
		},
	}
}
