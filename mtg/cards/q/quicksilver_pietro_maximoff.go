package q

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// QuicksilverPietroMaximoff is the card definition for Quicksilver, Pietro Maximoff.
//
// Type: Legendary Creature — Mutant Hero
// Cost: {2}{R}
//
// Oracle text:
//
//	Haste (This creature can attack and {T} as soon as he comes under your control.)
var QuicksilverPietroMaximoff = newQuicksilverPietroMaximoff

func newQuicksilverPietroMaximoff() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Quicksilver, Pietro Maximoff",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Mutant, types.Hero},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
			OracleText: `
			Haste (This creature can attack and {T} as soon as he comes under your control.)
		`,
		},
	}
}
