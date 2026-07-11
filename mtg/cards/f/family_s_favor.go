package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FamilySFavor is the card definition for Family's Favor.
//
// Type: Enchantment
// Cost: {2}{G}
//
// Oracle text:
//
//	Whenever you attack, put a shield counter on target attacking creature. Until end of turn, it gains "Whenever this creature deals combat damage to a player, remove a shield counter from it. If you do, draw a card." (If a creature with a shield counter on it would be dealt damage or destroyed, remove a shield counter from it instead.)
var FamilySFavor = newFamilySFavor

func newFamilySFavor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Family's Favor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:      game.EventAttackerDeclared,
							Controller: game.TriggerControllerYou,
							OneOrMore:  true,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target attacking creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.Shield,
								},
							},
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddAbilities: []game.Ability{
												new(game.TriggeredAbility{
													Trigger: game.TriggerCondition{
														Type: game.TriggerWhenever,
														Pattern: game.TriggerPattern{
															Event:               game.EventDamageDealt,
															Source:              game.TriggerSourceSelf,
															Subject:             game.TriggerSubjectDamageSource,
															RequireCombatDamage: true,
															DamageRecipient:     game.DamageRecipientPlayer,
														},
													},
													Content: game.Mode{
														Sequence: []game.Instruction{
															{
																Primitive: game.RemoveCounter{
																	Amount:      game.Fixed(1),
																	Object:      game.EventPermanentReference(),
																	CounterKind: counter.Shield,
																},
																PublishResult: game.ResultKey("if-you-do"),
															},
															{
																Primitive: game.Draw{
																	Amount: game.Fixed(1),
																	Player: game.ControllerReference(),
																},
																ResultGate: opt.Val(game.InstructionResultGate{
																	Key:       "if-you-do",
																	Succeeded: game.TriTrue,
																}),
															},
														},
													}.Ability(),
												}),
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever you attack, put a shield counter on target attacking creature. Until end of turn, it gains "Whenever this creature deals combat damage to a player, remove a shield counter from it. If you do, draw a card." (If a creature with a shield counter on it would be dealt damage or destroyed, remove a shield counter from it instead.)
		`,
		},
	}
}
