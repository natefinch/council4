package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SimicBasilisk is the card definition for Simic Basilisk.
//
// Type: Creature — Basilisk Mutant
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	Graft 3 (This creature enters with three +1/+1 counters on it. Whenever another creature enters, you may move a +1/+1 counter from this creature onto it.)
//	{1}{G}: Until end of turn, target creature with a +1/+1 counter on it gains "Whenever this creature deals combat damage to a creature, destroy that creature at end of combat."
var SimicBasilisk = newSimicBasilisk

func newSimicBasilisk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Simic Basilisk",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Basilisk, types.Mutant},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{1}{G}: Until end of turn, target creature with a +1/+1 counter on it gains \"Whenever this creature deals combat damage to a creature, destroy that creature at end of combat.\"",
					ManaCost:       opt.Val(cost.Mana{cost.O(1), cost.G}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature with a +1/+1 counter on it",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, MatchCounter: true, RequiredCounter: counter.PlusOnePlusOne}),
							},
						},
						Sequence: []game.Instruction{
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
															Event:                game.EventDamageDealt,
															Source:               game.TriggerSourceSelf,
															Subject:              game.TriggerSubjectDamageSource,
															RequireCombatDamage:  true,
															DamageRecipient:      game.DamageRecipientPermanent,
															DamageRecipientTypes: []types.Card{types.Creature},
														},
													},
													Content: game.Mode{
														Sequence: []game.Instruction{
															{
																Primitive: game.CreateDelayedTrigger{
																	Trigger: game.DelayedTriggerDef{
																		Timing:         game.DelayedAtEndOfCombat,
																		CapturedObject: opt.Val(game.EventPermanentReference()),
																		Content: game.Mode{
																			Sequence: []game.Instruction{
																				{
																					Primitive: game.Destroy{
																						Object: game.CapturedObjectReference(),
																					},
																				},
																			},
																		}.Ability(),
																	},
																},
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							ExcludeSelf:      true,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Optional: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCounters{
									Amount:      game.Fixed(1),
									Object:      game.EventPermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
									Source:      game.CounterSourceSpec{Kind: game.CounterSourceSelf},
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with three +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 3}),
			},
			OracleText: `
			Graft 3 (This creature enters with three +1/+1 counters on it. Whenever another creature enters, you may move a +1/+1 counter from this creature onto it.)
			{1}{G}: Until end of turn, target creature with a +1/+1 counter on it gains "Whenever this creature deals combat damage to a creature, destroy that creature at end of combat."
		`,
		},
	}
}
