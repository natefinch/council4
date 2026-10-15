package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Zephid is the card definition for Zephid.
//
// Type: Creature — Illusion
// Cost: {4}{U}{U}
//
// Oracle text:
//
//	Flying
//	Shroud (This creature can't be the target of spells or abilities.)
var Zephid = newZephid

func newZephid() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Zephid",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Illusion},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.ShroudStaticBody,
			},
			OracleText: `
			Flying
			Shroud (This creature can't be the target of spells or abilities.)
		`,
		},
	}
}
