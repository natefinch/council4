package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpitefulSliver is the card definition for Spiteful Sliver.
//
// Type: Creature — Sliver
// Cost: {2}{R}
//
// Oracle text:
//
//	Sliver creatures you control have "Whenever this creature is dealt damage, it deals that much damage to target player or planeswalker."
var SpitefulSliver = newSpitefulSliver

func newSpitefulSliver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Spiteful Sliver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sliver},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Sliver creatures you control have \"Whenever this creature is dealt damage, it deals that much damage to target player or planeswalker.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Sliver")}}),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.TriggeredAbility{
									Text: "Whenever this creature is dealt damage, it deals that much damage to target player or planeswalker.",
									Trigger: game.TriggerCondition{
										Type: game.TriggerWhenever,
										Pattern: game.TriggerPattern{
											Event:           game.EventDamageDealt,
											Source:          game.TriggerSourceSelf,
											Subject:         game.TriggerSubjectPermanent,
											StackObjectKind: game.StackSpell,
											DamageRecipient: game.DamageRecipientPermanent,
											CounterKind:     counter.PlusOnePlusOne,
										},
									},
									Content: game.AbilityContent{
										Modes: []game.Mode{
											game.Mode{
												Targets: []game.TargetSpec{
													game.TargetSpec{
														MinTargets: 1,
														MaxTargets: 1,
														Constraint: "target player or planeswalker",
														Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
														Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Planeswalker}}),
													},
												},
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.Damage{
															Amount: game.Dynamic(game.DynamicAmount{
																Kind:       game.DynamicAmountEventDamage,
																Multiplier: 1,
															}),
															Recipient:    game.AnyTargetDamageRecipient(0),
															DamageSource: opt.Val(game.EventPermanentReference()),
														},
													},
												},
											},
										},
										MinModes: 1,
										MaxModes: 1,
									},
								},
							},
						},
					},
				},
			},
			OracleText: `
			Sliver creatures you control have "Whenever this creature is dealt damage, it deals that much damage to target player or planeswalker."
		`,
		},
	}
}
