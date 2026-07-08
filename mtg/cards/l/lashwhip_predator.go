package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LashwhipPredator is the card definition for Lashwhip Predator.
//
// Type: Creature — Plant Beast
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	This spell costs {2} less to cast if your opponents control three or more creatures.
//	Reach
var LashwhipPredator = newLashwhipPredator

func newLashwhipPredator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Lashwhip Predator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Plant, types.Beast},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 2,
								ReductionCondition: opt.Val(game.Condition{
									OpponentsControl: opt.Val(game.SelectionCount{
										Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
										MinCount:  3,
									}),
								}),
							},
						},
					},
				},
				game.ReachStaticBody,
			},
			OracleText: `
			This spell costs {2} less to cast if your opponents control three or more creatures.
			Reach
		`,
		},
	}
}
