package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RegalSliver is the card definition for Regal Sliver.
//
// Type: Creature — Sliver
// Cost: {3}{W}
//
// Oracle text:
//
//	Sliver creatures you control have "When this creature enters, Slivers you control get +1/+1 until end of turn if you're the monarch. Otherwise, you become the monarch."
var RegalSliver = newRegalSliver()

func newRegalSliver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Regal Sliver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sliver},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Sliver")}}),
							AddAbilities: []game.Ability{
								new(game.TriggeredAbility{
									Trigger: game.TriggerCondition{
										Type: game.TriggerWhen,
										Pattern: game.TriggerPattern{
											Event:  game.EventPermanentEnteredBattlefield,
											Source: game.TriggerSourceSelf,
										},
									},
									Content: game.Mode{
										Sequence: []game.Instruction{
											{
												Primitive: game.ApplyContinuous{
													ContinuousEffects: []game.ContinuousEffect{
														game.ContinuousEffect{
															Layer:          game.LayerPowerToughnessModify,
															Group:          game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Sliver")}, Controller: game.ControllerYou}),
															PowerDelta:     1,
															ToughnessDelta: 1,
														},
													},
													Duration: game.DurationUntilEndOfTurn,
												},
												Condition: opt.Val(game.EffectCondition{
													Condition: opt.Val(game.Condition{
														ControllerIsMonarch: true,
													}),
												}),
											},
											{
												Primitive: game.BecomeMonarch{
													Player: game.ControllerReference(),
												},
												Condition: opt.Val(game.EffectCondition{
													Condition: opt.Val(game.Condition{
														Negate:              true,
														ControllerIsMonarch: true,
													}),
												}),
											},
										},
									}.Ability(),
								}),
							},
						},
					},
				},
			},
			OracleText: `
			Sliver creatures you control have "When this creature enters, Slivers you control get +1/+1 until end of turn if you're the monarch. Otherwise, you become the monarch."
		`,
		},
	}
}
