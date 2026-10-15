package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RiptideTurtle is the card definition for Riptide Turtle.
//
// Type: Creature — Turtle
// Cost: {1}{U}
//
// Oracle text:
//
//	Flash
//	Defender
var RiptideTurtle = newRiptideTurtle

func newRiptideTurtle() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Riptide Turtle",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Turtle},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.DefenderStaticBody,
			},
			OracleText: `
			Flash
			Defender
		`,
		},
	}
}
