package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FlameshadowConjuring is the card definition for Flameshadow Conjuring.
//
// Type: Enchantment
// Cost: {3}{R}
//
// Oracle text:
//
//	Whenever a nontoken creature you control enters, you may pay {R}. If you do, create a token that's a copy of that creature. That token gains haste. Exile it at the beginning of the next end step.
var FlameshadowConjuring = newFlameshadowConjuring

func newFlameshadowConjuring() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Flameshadow Conjuring",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever a nontoken creature you control enters, you may pay {R}. If you do, create a token that's a copy of that creature. That token gains haste. Exile it at the beginning of the next end step.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, NonToken: true},
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Pay{
											Payment: game.ResolutionPayment{
												Prompt: "Pay {R}?",
												ManaCost: opt.Val(cost.Mana{
													cost.R,
												}),
											},
										},
										PublishResult: game.ResultKey("controller-paid"),
									},
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenCopyOf(game.TokenCopySpec{
												Source: game.TokenCopySourceObject,
												Object: game.EventPermanentReference(),
											}),
											PublishLinked: game.LinkedKey("flameshadow-copy"),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:       game.ResultKey("controller-paid"),
											Succeeded: game.TriTrue,
										}),
									},
									game.Instruction{
										Primitive: game.ApplyContinuous{
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Group:       game.LinkedObjectsGroup(game.LinkedKey("flameshadow-copy")),
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.Haste},
												},
											},
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:       game.ResultKey("controller-paid"),
											Succeeded: game.TriTrue,
										}),
									},
									game.Instruction{
										Primitive: game.CreateDelayedTrigger{
											Trigger: game.DelayedTriggerDef{
												Timing: game.DelayedAtBeginningOfNextEndStep,
												Content: game.AbilityContent{
													Modes: []game.Mode{
														game.Mode{
															Sequence: []game.Instruction{
																game.Instruction{
																	Primitive: game.MovePermanent{
																		Group:       game.CapturedObjectsGroup(),
																		Destination: zone.Exile,
																	},
																},
															},
														},
													},
													MinModes: 1,
													MaxModes: 1,
												},
												CapturedObjectGroup: opt.Val(game.LinkedObjectReference("flameshadow-copy")),
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
			Whenever a nontoken creature you control enters, you may pay {R}. If you do, create a token that's a copy of that creature. That token gains haste. Exile it at the beginning of the next end step.
		`,
		},
	}
}
