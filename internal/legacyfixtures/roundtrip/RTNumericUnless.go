// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RTNumericUnless is the card definition for RT Numeric Unless.
//
// Type: Creature
//
// Oracle text:
//
//	When this creature enters, draw a card. Then discard a card unless you control another creature with power 2 or less.
var RTNumericUnless = newRTNumericUnless

func newRTNumericUnless() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "RT Numeric Unless",
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, draw a card. Then discard a card unless you control another creature with power 2 or less.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
									game.Instruction{
										Primitive: game.Discard{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:   "unless you control another creature with power 2 or less",
												Negate: true,
												ControlsMatching: opt.Val(game.SelectionCount{
													Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludeSource: true, Power: opt.Val(compare.Int{
														Op:    compare.LessOrEqual,
														Value: 2,
													})},
												}),
												SourceCounterKind: counter.PlusOnePlusOne,
											}),
										}),
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
			When this creature enters, draw a card. Then discard a card unless you control another creature with power 2 or less.
		`,
		},
	}
}
