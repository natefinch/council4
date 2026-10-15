package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WatcherSliver is the card definition for Watcher Sliver.
//
// Type: Creature — Sliver
// Cost: {3}{W}
//
// Oracle text:
//
//	All Sliver creatures get +0/+2.
var WatcherSliver = newWatcherSliver

func newWatcherSliver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Watcher Sliver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sliver},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "All Sliver creatures get +0/+2.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Sliver")}}),
							Layer:          game.LayerPowerToughnessModify,
							ToughnessDelta: 2,
						},
					},
				},
			},
			OracleText: `
			All Sliver creatures get +0/+2.
		`,
		},
	}
}
