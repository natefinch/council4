package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MoggFlunkies is the card definition for Mogg Flunkies.
//
// Type: Creature — Goblin
// Cost: {1}{R}
//
// Oracle text:
//
//	This creature can't attack or block alone.
var MoggFlunkies = newMoggFlunkies()

func newMoggFlunkies() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Mogg Flunkies",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.CantAttackOrBlockAloneStaticBody,
			},
			OracleText: `
			This creature can't attack or block alone.
		`,
		},
	}
}
