package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MinscBooTimelessHeroes is the card definition for Minsc & Boo, Timeless Heroes.
//
// Type: Legendary Planeswalker — Minsc
// Cost: {2}{R}{G}
//
// Oracle text:
//
//	When Minsc & Boo enters and at the beginning of your upkeep, you may create Boo, a legendary 1/1 red Hamster creature token with trample and haste.
//	+1: Put three +1/+1 counters on up to one target creature with trample or haste.
//	−2: Sacrifice a creature. When you do, Minsc & Boo deals X damage to any target, where X is that creature's power. If the sacrificed creature was a Hamster, draw X cards.
//	Minsc & Boo, Timeless Heroes can be your commander.
var MinscBooTimelessHeroes = newMinscBooTimelessHeroes

func newMinscBooTimelessHeroes() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Minsc & Boo, Timeless Heroes",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.G,
			}),
			Colors:         []color.Color{color.Green, color.Red},
			CanBeCommander: true,
			Supertypes:     []types.Super{types.Legendary},
			Types:          []types.Card{types.Planeswalker},
			Subtypes:       []types.Sub{types.Minsc},
			Loyalty:        opt.Val(3),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Minsc & Boo enters, you may create Boo, a legendary 1/1 red Hamster creature token with trample and haste.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Optional: true,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(minscBooTimelessHeroesToken),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.TriggeredAbility{
					Text: "At the beginning of your upkeep, you may create Boo, a legendary 1/1 red Hamster creature token with trample and haste.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepUpkeep,
						},
					},
					Optional: true,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(minscBooTimelessHeroesToken),
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
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					Text:        "+1: Put three +1/+1 counters on up to one target creature with trample or haste.",
					LoyaltyCost: 1,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MaxTargets: 1,
										Constraint: "up to one target creature with trample or haste",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{AnyOf: []game.Selection{game.Selection{Keyword: game.Trample}, game.Selection{Keyword: game.Haste}}, RequiredTypesAny: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(3),
											Object:      game.TargetPermanentReference(0),
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
				game.LoyaltyAbility{
					Text:        "−2: Sacrifice a creature. When you do, Minsc & Boo deals X damage to any target, where X is that creature's power. If the sacrificed creature was a Hamster, draw X cards.",
					LoyaltyCost: -2,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.SacrificePermanents{
											Player:               game.ControllerReference(),
											Amount:               game.Fixed(1),
											Selection:            game.Selection{RequiredTypes: []types.Card{types.Creature}},
											PublishLinked:        game.LinkedKey("sacrificed-creature"),
											PublishObjectBinding: true,
										},
										PublishResult: game.ResultKey("if-you-do"),
									},
									game.Instruction{
										Primitive: game.CreateReflexiveTrigger{
											Trigger: game.ReflexiveTriggerDef{
												Content: game.AbilityContent{
													Modes: []game.Mode{
														game.Mode{
															Targets: []game.TargetSpec{
																game.TargetSpec{
																	MinTargets: 1,
																	MaxTargets: 1,
																	Constraint: "any target",
																	Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
																},
															},
															Sequence: []game.Instruction{
																game.Instruction{
																	Primitive: game.Damage{
																		Amount: game.Dynamic(game.DynamicAmount{
																			Kind:       game.DynamicAmountObjectPower,
																			Multiplier: 1,
																			Object:     game.LinkedObjectReference("sacrificed-creature"),
																		}),
																		Recipient:    game.AnyTargetDamageRecipient(0),
																		DamageSource: opt.Val(game.SourcePermanentReference()),
																	},
																},
																game.Instruction{
																	Primitive: game.Draw{
																		Amount: game.Dynamic(game.DynamicAmount{
																			Kind:       game.DynamicAmountObjectPower,
																			Multiplier: 1,
																			Object:     game.LinkedObjectReference("sacrificed-creature"),
																		}),
																		Player: game.ControllerReference(),
																	},
																	Condition: opt.Val(game.EffectCondition{
																		Text:   "If the sacrificed creature was a Hamster",
																		Object: game.LinkedObjectReference("sacrificed-creature"),
																		Condition: opt.Val(game.Condition{
																			Text:              "If the sacrificed creature was a Hamster",
																			Object:            opt.Val(game.LinkedObjectReference("sacrificed-creature")),
																			ObjectMatches:     opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Hamster}}),
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
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:       game.ResultKey("if-you-do"),
											Succeeded: game.TriTrue,
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
			When Minsc & Boo enters and at the beginning of your upkeep, you may create Boo, a legendary 1/1 red Hamster creature token with trample and haste.
			+1: Put three +1/+1 counters on up to one target creature with trample or haste.
			−2: Sacrifice a creature. When you do, Minsc & Boo deals X damage to any target, where X is that creature's power. If the sacrificed creature was a Hamster, draw X cards.
			Minsc & Boo, Timeless Heroes can be your commander.
		`,
		},
	}
}

var minscBooTimelessHeroesToken = newMinscBooTimelessHeroesToken()

func newMinscBooTimelessHeroesToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:       "Boo",
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Hamster},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.HasteStaticBody,
			},
		},
	}
}
