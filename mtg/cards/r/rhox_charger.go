package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RhoxCharger is the card definition for Rhox Charger.
//
// Type: Creature — Rhino Soldier
// Cost: {3}{G}
//
// Oracle text:
//
//	Trample
//	Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
var RhoxCharger = newRhoxCharger

func newRhoxCharger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Rhox Charger",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Rhino, types.Soldier},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.ExaltedStaticBody,
			},
			OracleText: `
			Trample
			Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
		`,
		},
	}
}
