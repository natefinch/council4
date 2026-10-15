package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BattleSliver is the card definition for Battle Sliver.
//
// Type: Creature — Sliver
// Cost: {4}{R}
//
// Oracle text:
//
//	Sliver creatures you control get +2/+0.
var BattleSliver = newBattleSliver

func newBattleSliver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Battle Sliver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sliver},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Sliver creatures you control get +2/+0.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Sliver")}}),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 2,
						},
					},
				},
			},
			OracleText: `
			Sliver creatures you control get +2/+0.
		`,
		},
	}
}
