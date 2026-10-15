package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KederektCreeper is the card definition for Kederekt Creeper.
//
// Type: Creature — Horror
// Cost: {U}{B}{R}
//
// Oracle text:
//
//	Menace (This creature can't be blocked except by two or more creatures.)
//	Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)
var KederektCreeper = newKederektCreeper

func newKederektCreeper() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Kederekt Creeper",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.B,
				cost.R,
			}),
			Colors:    []color.Color{color.Black, color.Red, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Horror},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
				game.DeathtouchStaticBody,
			},
			OracleText: `
			Menace (This creature can't be blocked except by two or more creatures.)
			Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)
		`,
		},
	}
}
