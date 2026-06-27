package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SliverLegion is the card definition for Sliver Legion.
//
// Type: Legendary Creature — Sliver
// Cost: {W}{U}{B}{R}{G}
//
// Oracle text:
//
//	All Sliver creatures get +1/+1 for each other Sliver on the battlefield.
var SliverLegion = newSliverLegion()

func newSliverLegion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Sliver Legion",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.U,
				cost.B,
				cost.R,
				cost.G,
			}),
			Colors:     []color.Color{color.Black, color.Green, color.Red, color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Sliver},
			Power:      opt.Val(game.PT{Value: 7}),
			Toughness:  opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerPowerToughnessModify,
							Group: game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Sliver")}}),
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Sliver")}, ExcludeSource: true}),
							}),
							ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Sliver")}, ExcludeSource: true}),
							}),
						},
					},
				},
			},
			OracleText: `
			All Sliver creatures get +1/+1 for each other Sliver on the battlefield.
		`,
		},
	}
}
