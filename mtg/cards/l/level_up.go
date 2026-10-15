package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LevelUp is the card definition for Level Up.
//
// Type: Enchantment — Aura
// Cost: {1}{G}
//
// Oracle text:
//
//	Enchant creature
//	When this Aura enters, put a +1/+1 counter on enchanted creature.
//	Enchanted creature has "Whenever this creature attacks, double the number of +1/+1 counters on it. Then if it has power 10 or greater, draw a card."
var LevelUp = newLevelUp

func newLevelUp() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Level Up",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					Text: "Enchanted creature has \"Whenever this creature attacks, double the number of +1/+1 counters on it. Then if it has power 10 or greater, draw a card.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.TriggeredAbility{
									Text: "Whenever this creature attacks, double the number of +1/+1 counters on it. Then if it has power 10 or greater, draw a card.",
									Trigger: game.TriggerCondition{
										Type: game.TriggerWhenever,
										Pattern: game.TriggerPattern{
											Event:           game.EventAttackerDeclared,
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
														Primitive: game.AddCounter{
															Amount: game.Dynamic(game.DynamicAmount{
																Kind:        game.DynamicAmountObjectCounters,
																CounterKind: counter.PlusOnePlusOne,
																Object:      game.EventPermanentReference(),
															}),
															Object:      game.EventPermanentReference(),
															CounterKind: counter.PlusOnePlusOne,
														},
													},
													game.Instruction{
														Primitive: game.Draw{
															Amount: game.Fixed(1),
															Player: game.ControllerReference(),
														},
														Condition: opt.Val(game.EffectCondition{
															Condition: opt.Val(game.Condition{
																Text:   "if it has power 10 or greater",
																Object: opt.Val(game.EventPermanentReference()),
																ObjectMatches: opt.Val(game.Selection{Power: opt.Val(compare.Int{
																	Op:    compare.GreaterOrEqual,
																	Value: 10,
																})}),
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
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this Aura enters, put a +1/+1 counter on enchanted creature.",
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
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourceAttachedPermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
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
			OracleText: `
			Enchant creature
			When this Aura enters, put a +1/+1 counter on enchanted creature.
			Enchanted creature has "Whenever this creature attacks, double the number of +1/+1 counters on it. Then if it has power 10 or greater, draw a card."
		`,
		},
	}
}
