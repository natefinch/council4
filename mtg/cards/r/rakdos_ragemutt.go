package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RakdosRagemutt is the card definition for Rakdos Ragemutt.
//
// Type: Creature — Elemental Dog
// Cost: {3}{B}{R}
//
// Oracle text:
//
//	Lifelink, haste
var RakdosRagemutt = newRakdosRagemutt

func newRakdosRagemutt() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Rakdos Ragemutt",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.R,
			}),
			Colors:    []color.Color{color.Black, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental, types.Dog},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.LifelinkStaticBody,
				game.HasteStaticBody,
			},
			OracleText: `
			Lifelink, haste
		`,
		},
	}
}
