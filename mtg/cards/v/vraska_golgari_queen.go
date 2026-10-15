package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VraskaGolgariQueen is the card definition for Vraska, Golgari Queen.
//
// Type: Legendary Planeswalker — Vraska
// Cost: {2}{B}{G}
//
// Oracle text:
//
//	+2: You may sacrifice another permanent. If you do, you gain 1 life and draw a card.
//	−3: Destroy target nonland permanent with mana value 3 or less.
//	−9: You get an emblem with "Whenever a creature you control deals combat damage to a player, that player loses the game."
var VraskaGolgariQueen = newVraskaGolgariQueen

func newVraskaGolgariQueen() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Vraska, Golgari Queen",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.G,
			}),
			Colors:     []color.Color{color.Black, color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Vraska},
			Loyalty:    opt.Val(4),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					Text:        "+2: You may sacrifice another permanent. If you do, you gain 1 life and draw a card.",
					LoyaltyCost: 2,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.SacrificePermanents{
											Player:    game.ControllerReference(),
											Amount:    game.Fixed(1),
											Selection: game.Selection{ExcludeSource: true},
										},
										Optional:      true,
										PublishResult: game.ResultKey("if-you-do"),
										LocalProducts: game.LocalProducts{
											Results: []game.ResultKey{game.ResultKey("if-you-do")},
										},
									},
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:       game.ResultKey("if-you-do"),
											Succeeded: game.TriTrue,
										}),
									},
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
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
					Text:        "−3: Destroy target nonland permanent with mana value 3 or less.",
					LoyaltyCost: -3,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target nonland permanent with mana value 3 or less",
										Allow:      game.TargetAllowPermanent,
										Selection: opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}, ManaValue: opt.Val(compare.Int{
											Op:    compare.LessOrEqual,
											Value: 3,
										})}),
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
					Text:        "−9: You get an emblem with \"Whenever a creature you control deals combat damage to a player, that player loses the game.\"",
					LoyaltyCost: -9,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateEmblem{
											EmblemAbilities: []game.Ability{
												&game.TriggeredAbility{
													Text: "Whenever a creature you control deals combat damage to a player, that player loses the game.",
													Trigger: game.TriggerCondition{
														Type: game.TriggerWhenever,
														Pattern: game.TriggerPattern{
															Event:                 game.EventDamageDealt,
															Controller:            game.TriggerControllerYou,
															Subject:               game.TriggerSubjectDamageSource,
															StackObjectKind:       game.StackSpell,
															DamageRecipient:       game.DamageRecipientPlayer,
															DamageSourceSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
															RequireCombatDamage:   true,
															CounterKind:           counter.PlusOnePlusOne,
														},
													},
													Content: game.AbilityContent{
														Modes: []game.Mode{
															game.Mode{
																Sequence: []game.Instruction{
																	game.Instruction{
																		Primitive: game.PlayerLosesGame{
																			Player: game.EventPlayerReference(),
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
			+2: You may sacrifice another permanent. If you do, you gain 1 life and draw a card.
			−3: Destroy target nonland permanent with mana value 3 or less.
			−9: You get an emblem with "Whenever a creature you control deals combat damage to a player, that player loses the game."
		`,
		},
	}
}
