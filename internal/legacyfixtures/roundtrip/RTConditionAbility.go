// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RTConditionAbility is the card definition for RT Condition Ability.
//
// Type: Enchantment
//
// Oracle text:
//
//	{1}: If you have no cards in hand, draw a card, then draw a card.
var RTConditionAbility = newRTConditionAbility

func newRTConditionAbility() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Condition Ability",
			Types: []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}: If you have no cards in hand, draw a card, then draw a card.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:                "If you have no cards in hand",
												ControllerHandEmpty: true,
												SourceCounterKind:   counter.PlusOnePlusOne,
											}),
										}),
										PublishCondition: game.ConditionKey("condition-0"),
									},
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
										ConditionGate: game.ConditionKey("condition-0"),
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			{1}: If you have no cards in hand, draw a card, then draw a card.
		`,
		},
	}
}
