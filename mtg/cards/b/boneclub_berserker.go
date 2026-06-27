package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BoneclubBerserker is the card definition for Boneclub Berserker.
//
// Type: Creature — Goblin Berserker
// Cost: {3}{R}
//
// Oracle text:
//
//	This creature gets +2/+0 for each other Goblin you control.
var BoneclubBerserker = newBoneclubBerserker()

func newBoneclubBerserker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Boneclub Berserker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Berserker},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							AffectedSource: true,
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 2,
								Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Goblin")}, Controller: game.ControllerYou, ExcludeSource: true}),
							}),
						},
					},
				},
			},
			OracleText: `
			This creature gets +2/+0 for each other Goblin you control.
		`,
		},
	}
}
