package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GuideOfSouls is the card definition for Guide of Souls.
//
// Type: Creature — Human Cleric
// Cost: {W}
//
// Oracle text:
//
//	Whenever another creature you control enters, you gain 1 life and get {E} (an energy counter).
//	Whenever you attack, you may pay {E}{E}{E}. When you do, put two +1/+1 counters and a flying counter on target attacking creature. It becomes an Angel in addition to its other types.
var GuideOfSouls = newGuideOfSouls

func newGuideOfSouls() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Guide of Souls",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Cleric},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever another creature you control enters, you gain 1 life and get {E} (an energy counter).",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							ExcludeSelf:      true,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
									game.Instruction{
										Primitive: game.AddPlayerCounter{
											Amount:      game.Fixed(1),
											Player:      game.ControllerReference(),
											CounterKind: counter.Energy,
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
					Text: "Whenever you attack, you may pay {E}{E}{E}. When you do, put two +1/+1 counters and a flying counter on target attacking creature. It becomes an Angel in addition to its other types.",
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
												Prompt: "Pay {E}{E}{E}?",
												AdditionalCosts: []cost.Additional{
													cost.Additional{
														Kind:        cost.AdditionalEnergy,
														Text:        "pay {E}{E}{E}",
														Amount:      3,
														CounterKind: counter.PlusOnePlusOne,
													},
												},
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
																	Constraint: "target attacking creature",
																	Allow:      game.TargetAllowPermanent,
																	Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}),
																},
															},
															Sequence: []game.Instruction{
																game.Instruction{
																	Primitive: game.AddCounter{
																		Amount:      game.Fixed(2),
																		Object:      game.TargetPermanentReference(0),
																		CounterKind: counter.PlusOnePlusOne,
																	},
																},
																game.Instruction{
																	Primitive: game.AddCounter{
																		Amount:      game.Fixed(1),
																		Object:      game.TargetPermanentReference(0),
																		CounterKind: counter.Flying,
																	},
																},
																game.Instruction{
																	Primitive: game.ApplyContinuous{
																		Object: opt.Val(game.TargetPermanentReference(0)),
																		ContinuousEffects: []game.ContinuousEffect{
																			game.ContinuousEffect{
																				Duration:    game.DurationPermanent,
																				Layer:       game.LayerType,
																				AddSubtypes: []types.Sub{types.Angel},
																			},
																		},
																		Duration: game.DurationPermanent,
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
			Whenever another creature you control enters, you gain 1 life and get {E} (an energy counter).
			Whenever you attack, you may pay {E}{E}{E}. When you do, put two +1/+1 counters and a flying counter on target attacking creature. It becomes an Angel in addition to its other types.
		`,
		},
	}
}
