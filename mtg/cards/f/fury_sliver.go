package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FurySliver is the card definition for Fury Sliver.
//
// Type: Creature — Sliver
// Cost: {5}{R}
//
// Oracle text:
//
//	All Sliver creatures have double strike.
var FurySliver = newFurySliver

func newFurySliver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Fury Sliver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sliver},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "All Sliver creatures have double strike.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Sliver")}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.DoubleStrike},
						},
					},
				},
			},
			OracleText: `
			All Sliver creatures have double strike.
		`,
		},
	}
}
