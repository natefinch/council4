package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// InvisibleWoman is the card definition for Invisible Woman.
//
// Type: Legendary Creature — Human Hero
// Cost: {2}{W}
//
// Oracle text:
//
//	At the beginning of combat on your turn, if you've cast a noncreature spell this turn, create a 0/3 colorless Wall creature token with defender and reach.
//	Whenever you attack, you may pay {R}{G}{W}{U}. When you do, target creature gets +1/+0 until end of turn for each creature you control and can't be blocked this turn.
var InvisibleWoman = newInvisibleWoman

func newInvisibleWoman() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Invisible Woman",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Hero},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of combat on your turn, if you've cast a noncreature spell this turn, create a 0/3 colorless Wall creature token with defender and reach.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepBeginningOfCombat,
						},
						InterveningIf: "if you've cast a noncreature spell this turn",
						InterveningCondition: opt.Val(game.Condition{
							Text:              "if you've cast a noncreature spell this turn",
							SourceCounterKind: counter.PlusOnePlusOne,
							EventHistory: opt.Val(game.EventHistoryCondition{
								Pattern: game.TriggerPattern{
									Event:           game.EventSpellCast,
									Controller:      game.TriggerControllerYou,
									CardSelection:   game.Selection{ExcludedTypes: []types.Card{types.Creature}},
									StackObjectKind: game.StackSpell,
									CounterKind:     counter.PlusOnePlusOne,
								},
								Window: game.EventHistoryCurrentTurn,
							}),
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(invisibleWomanToken),
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
					Text: "Whenever you attack, you may pay {R}{G}{W}{U}. When you do, target creature gets +1/+0 until end of turn for each creature you control and can't be blocked this turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerDeclared,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							OneOrMore:       true,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Pay{
											Payment: game.ResolutionPayment{
												Prompt: "Pay {R}{G}{W}{U}?",
												ManaCost: opt.Val(cost.Mana{
													cost.R,
													cost.G,
													cost.W,
													cost.U,
												}),
											},
										},
										PublishResult: game.ResultKey("controller-paid"),
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
																	Constraint: "target creature",
																	Allow:      game.TargetAllowPermanent,
																	Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
																},
															},
															Sequence: []game.Instruction{
																game.Instruction{
																	Primitive: game.ModifyPT{
																		Object: game.TargetPermanentReference(0),
																		PowerDelta: game.Dynamic(game.DynamicAmount{
																			Kind:       game.DynamicAmountCountSelector,
																			Multiplier: 1,
																			Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
																		}),
																		Duration: game.DurationUntilEndOfTurn,
																	},
																},
																game.Instruction{
																	Primitive: game.ApplyRule{
																		Object: opt.Val(game.TargetPermanentReference(0)),
																		RuleEffects: []game.RuleEffect{
																			game.RuleEffect{
																				Kind:     game.RuleEffectCantBeBlocked,
																				Duration: game.DurationPermanent,
																			},
																		},
																		Duration: game.DurationThisTurn,
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
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:       game.ResultKey("controller-paid"),
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
			At the beginning of combat on your turn, if you've cast a noncreature spell this turn, create a 0/3 colorless Wall creature token with defender and reach.
			Whenever you attack, you may pay {R}{G}{W}{U}. When you do, target creature gets +1/+0 until end of turn for each creature you control and can't be blocked this turn.
		`,
		},
	}
}

var invisibleWomanToken = newInvisibleWomanToken()

func newInvisibleWomanToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Wall",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
				game.ReachStaticBody,
			},
		},
	}
}
