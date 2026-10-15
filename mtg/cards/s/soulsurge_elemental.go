package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SoulsurgeElemental is the card definition for Soulsurge Elemental.
//
// Type: Creature — Elemental
// Cost: {3}{R}
//
// Oracle text:
//
//	First strike
//	Soulsurge Elemental's power is equal to the number of creatures you control.
var SoulsurgeElemental = newSoulsurgeElemental

func newSoulsurgeElemental() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Soulsurge Elemental",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:       []color.Color{color.Red},
			Types:        []types.Card{types.Creature},
			Subtypes:     []types.Sub{types.Elemental},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 1}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCount}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
			},
			OracleText: `
			First strike
			Soulsurge Elemental's power is equal to the number of creatures you control.
		`,
		},
	}
}
