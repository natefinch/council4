package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DarettiScrapSavant is the card definition for Daretti, Scrap Savant.
//
// Type: Legendary Planeswalker — Daretti
// Cost: {3}{R}
//
// Oracle text:
//
//	+2: Discard up to two cards, then draw that many cards.
//	−2: Sacrifice an artifact. If you do, return target artifact card from your graveyard to the battlefield.
//	−10: You get an emblem with "Whenever an artifact is put into your graveyard from the battlefield, return that card to the battlefield at the beginning of the next end step."
//	Daretti, Scrap Savant can be your commander.
var DarettiScrapSavant = newDarettiScrapSavant

func newDarettiScrapSavant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Daretti, Scrap Savant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:         []color.Color{color.Red},
			CanBeCommander: true,
			Supertypes:     []types.Super{types.Legendary},
			Types:          []types.Card{types.Planeswalker},
			Subtypes:       []types.Sub{types.Daretti},
			Loyalty:        opt.Val(3),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					Text:        "+2: Discard up to two cards, then draw that many cards.",
					LoyaltyCost: 2,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.DiscardThenDraw{
											Player: game.ControllerReference(),
											Max:    2,
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
					Text:        "−2: Sacrifice an artifact. If you do, return target artifact card from your graveyard to the battlefield.",
					LoyaltyCost: -2,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target artifact card from your graveyard",
										Allow:      game.TargetAllowCard,
										Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact}, Controller: game.ControllerYou}),
										TargetZone: zone.Graveyard,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.SacrificePermanents{
											Player:    game.ControllerReference(),
											Amount:    game.Fixed(1),
											Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
										},
										PublishResult: game.ResultKey("if-you-do"),
										LocalProducts: game.LocalProducts{
											Results: []game.ResultKey{game.ResultKey("if-you-do")},
										},
									},
									game.Instruction{
										Primitive: game.PutOnBattlefield{
											Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
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
				game.LoyaltyAbility{
					Text:        "−10: You get an emblem with \"Whenever an artifact is put into your graveyard from the battlefield, return that card to the battlefield at the beginning of the next end step.\"",
					LoyaltyCost: -10,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateEmblem{
											EmblemAbilities: []game.Ability{
												&game.TriggeredAbility{
													Text: "Whenever an artifact is put into your graveyard from the battlefield, return that card to the battlefield at the beginning of the next end step.",
													Trigger: game.TriggerCondition{
														Type: game.TriggerWhenever,
														Pattern: game.TriggerPattern{
															Event:            game.EventZoneChanged,
															Player:           game.TriggerPlayerYou,
															SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
															MatchFromZone:    true,
															FromZone:         zone.Battlefield,
															MatchToZone:      true,
															ToZone:           zone.Graveyard,
															StackObjectKind:  game.StackSpell,
															CounterKind:      counter.PlusOnePlusOne,
														},
													},
													Content: game.AbilityContent{
														Modes: []game.Mode{
															game.Mode{
																Sequence: []game.Instruction{
																	game.Instruction{
																		Primitive: game.CreateDelayedTrigger{
																			Trigger: game.DelayedTriggerDef{
																				Timing: game.DelayedAtBeginningOfNextEndStep,
																				Content: game.AbilityContent{
																					Modes: []game.Mode{
																						game.Mode{
																							Sequence: []game.Instruction{
																								game.Instruction{
																									Primitive: game.PutOnBattlefield{
																										Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceCaptured}),
																									},
																								},
																							},
																						},
																					},
																					MinModes: 1,
																					MaxModes: 1,
																				},
																				CapturedCard: opt.Val(game.EventPermanentReference()),
																			},
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
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			+2: Discard up to two cards, then draw that many cards.
			−2: Sacrifice an artifact. If you do, return target artifact card from your graveyard to the battlefield.
			−10: You get an emblem with "Whenever an artifact is put into your graveyard from the battlefield, return that card to the battlefield at the beginning of the next end step."
			Daretti, Scrap Savant can be your commander.
		`,
		},
	}
}
