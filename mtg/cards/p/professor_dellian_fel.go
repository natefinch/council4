package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ProfessorDellianFel is the card definition for Professor Dellian Fel.
//
// Type: Legendary Planeswalker — Dellian
// Cost: {2}{B}{G}
//
// Oracle text:
//
//	+2: You gain 3 life.
//	0: You draw a card and lose 1 life.
//	−3: Destroy target creature.
//	−6: You get an emblem with "Whenever you gain life, target opponent loses that much life."
var ProfessorDellianFel = newProfessorDellianFel

func newProfessorDellianFel() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Professor Dellian Fel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.G,
			}),
			Colors:     []color.Color{color.Black, color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Dellian},
			Loyalty:    opt.Val(5),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					Text:        "+2: You gain 3 life.",
					LoyaltyCost: 2,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(3),
											Player: game.ControllerReference(),
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
					Text: "0: You draw a card and lose 1 life.",
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
									game.Instruction{
										Primitive: game.LoseLife{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
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
					Text:        "−3: Destroy target creature.",
					LoyaltyCost: -3,
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
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(0),
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
					Text:        "−6: You get an emblem with \"Whenever you gain life, target opponent loses that much life.\"",
					LoyaltyCost: -6,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateEmblem{
											EmblemAbilities: []game.Ability{
												&game.TriggeredAbility{
													Text: "Whenever you gain life, target opponent loses that much life.",
													Trigger: game.TriggerCondition{
														Type: game.TriggerWhenever,
														Pattern: game.TriggerPattern{
															Event:           game.EventLifeGained,
															Player:          game.TriggerPlayerYou,
															StackObjectKind: game.StackSpell,
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
																		Constraint: "target opponent",
																		Allow:      game.TargetAllowPlayer,
																		Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
																	},
																},
																Sequence: []game.Instruction{
																	game.Instruction{
																		Primitive: game.LoseLife{
																			Amount: game.Dynamic(game.DynamicAmount{
																				Kind:       game.DynamicAmountEventLifeChange,
																				Multiplier: 1,
																			}),
																			Player: game.TargetPlayerReference(0),
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
			+2: You gain 3 life.
			0: You draw a card and lose 1 life.
			−3: Destroy target creature.
			−6: You get an emblem with "Whenever you gain life, target opponent loses that much life."
		`,
		},
	}
}
