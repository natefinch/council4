package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DapperShieldmate is the card definition for Dapper Shieldmate.
//
// Type: Creature — Human Soldier
// Cost: {3}{W}
//
// Oracle text:
//
//	This creature enters with a shield counter on it. (If it would be dealt damage or destroyed, remove a shield counter from it instead.)
//	During your turn, this creature gets +2/+0.
var DapperShieldmate = newDapperShieldmate

func newDapperShieldmate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Dapper Shieldmate",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "During your turn, this creature gets +2/+0.",
					Condition: opt.Val(game.Condition{
						Text:                 "During your turn",
						SourceCounterKind:    counter.PlusOnePlusOne,
						SourceControllerTurn: true,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with a shield counter on it. (If it would be dealt damage or destroyed, remove a shield counter from it instead.)", game.CounterPlacement{Kind: counter.Shield, Amount: 1}),
			},
			OracleText: `
			This creature enters with a shield counter on it. (If it would be dealt damage or destroyed, remove a shield counter from it instead.)
			During your turn, this creature gets +2/+0.
		`,
		},
	}
}
