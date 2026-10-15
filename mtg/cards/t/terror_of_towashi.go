package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TerrorOfTowashi is the card definition for Terror of Towashi.
//
// Type: Creature — Phyrexian Ogre
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	Deathtouch
//	Whenever this creature attacks, you may pay {3}{B}. When you do, return target creature card from your graveyard to the battlefield. It's a Phyrexian in addition to its other types.
var TerrorOfTowashi = newTerrorOfTowashi

func newTerrorOfTowashi() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Terror of Towashi",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Ogre},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DeathtouchStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature attacks, you may pay {3}{B}. When you do, return target creature card from your graveyard to the battlefield. It's a Phyrexian in addition to its other types.",
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
										Primitive: game.Pay{
											Payment: game.ResolutionPayment{
												Prompt: "Pay {3}{B}?",
												ManaCost: opt.Val(cost.Mana{
													cost.O(3),
													cost.B,
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
																	Constraint: "target creature card from your graveyard",
																	Allow:      game.TargetAllowCard,
																	Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
																	TargetZone: zone.Graveyard,
																},
															},
															Sequence: []game.Instruction{
																game.Instruction{
																	Primitive: game.PutOnBattlefield{
																		Source:        game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
																		PublishLinked: game.LinkedKey("sequence-effect-0-product"),
																	},
																	LocalProducts: game.LocalProducts{
																		Links: []game.LinkedKey{game.LinkedKey("sequence-effect-0-product")},
																	},
																},
																game.Instruction{
																	Primitive: game.ApplyContinuous{
																		Object: opt.Val(game.LinkedObjectReference("sequence-effect-0-product")),
																		ContinuousEffects: []game.ContinuousEffect{
																			game.ContinuousEffect{
																				Duration:    game.DurationPermanent,
																				Layer:       game.LayerType,
																				AddSubtypes: []types.Sub{types.Phyrexian},
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
			Deathtouch
			Whenever this creature attacks, you may pay {3}{B}. When you do, return target creature card from your graveyard to the battlefield. It's a Phyrexian in addition to its other types.
		`,
		},
	}
}
