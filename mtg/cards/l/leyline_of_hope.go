package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LeylineOfHope is the card definition for Leyline of Hope.
//
// Type: Enchantment
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	If this card is in your opening hand, you may begin the game with it on the battlefield.
//	If you would gain life, you gain that much life plus 1 instead.
//	As long as you have at least 7 life more than your starting life total, creatures you control get +2/+2.
var LeylineOfHope = newLeylineOfHope

func newLeylineOfHope() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Leyline of Hope",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					BeginsGameOnBattlefield: true,
				},
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerLifeAboveStarting, Op: compare.GreaterOrEqual, Value: 7}},
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.LifeGainReplacement("If you would gain life, you gain that much life plus 1 instead.", 1, 1),
			},
			OracleText: `
			If this card is in your opening hand, you may begin the game with it on the battlefield.
			If you would gain life, you gain that much life plus 1 instead.
			As long as you have at least 7 life more than your starting life total, creatures you control get +2/+2.
		`,
		},
	}
}
