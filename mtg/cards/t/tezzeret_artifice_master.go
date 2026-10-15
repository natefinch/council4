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

// TezzeretArtificeMaster is the card definition for Tezzeret, Artifice Master.
//
// Type: Legendary Planeswalker — Tezzeret
// Cost: {3}{U}{U}
//
// Oracle text:
//
//	+1: Create a 1/1 colorless Thopter artifact creature token with flying.
//	0: Draw a card. If you control three or more artifacts, draw two cards instead.
//	−9: You get an emblem with "At the beginning of your end step, search your library for a permanent card, put it onto the battlefield, then shuffle."
var TezzeretArtificeMaster = newTezzeretArtificeMaster

func newTezzeretArtificeMaster() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Tezzeret, Artifice Master",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Tezzeret},
			Loyalty:    opt.Val(5),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					Text:        "+1: Create a 1/1 colorless Thopter artifact creature token with flying.",
					LoyaltyCost: 1,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(tezzeretArtificeMasterToken),
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
					Text: "0: Draw a card. If you control three or more artifacts, draw two cards instead.",
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:   "If you control three or more artifacts",
												Negate: true,
												ControlsMatching: opt.Val(game.SelectionCount{
													Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
													MinCount:  3,
												}),
												SourceCounterKind: counter.PlusOnePlusOne,
											}),
										}),
										PublishCondition: game.ConditionKey("condition-0"),
									},
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(2),
											Player: game.ControllerReference(),
										},
										ConditionGate:       game.ConditionKey("condition-0"),
										ConditionGateNegate: true,
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.LoyaltyAbility{
					Text:        "−9: You get an emblem with \"At the beginning of your end step, search your library for a permanent card, put it onto the battlefield, then shuffle.\"",
					LoyaltyCost: -9,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateEmblem{
											EmblemAbilities: []game.Ability{
												&game.TriggeredAbility{
													Text: "At the beginning of your end step, search your library for a permanent card, put it onto the battlefield, then shuffle.",
													Trigger: game.TriggerCondition{
														Type: game.TriggerAt,
														Pattern: game.TriggerPattern{
															Event:           game.EventBeginningOfStep,
															Controller:      game.TriggerControllerYou,
															StackObjectKind: game.StackSpell,
															CounterKind:     counter.PlusOnePlusOne,
															Step:            game.StepEnd,
														},
													},
													Content: game.AbilityContent{
														Modes: []game.Mode{
															game.Mode{
																Sequence: []game.Instruction{
																	game.Instruction{
																		Primitive: game.Search{
																			Player: game.ControllerReference(),
																			Spec: game.SearchSpec{
																				SourceZone:  zone.Library,
																				Destination: zone.Battlefield,
																				Filter:      game.Selection{RequirePermanentCard: true},
																			},
																			Amount: game.Fixed(1),
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
			+1: Create a 1/1 colorless Thopter artifact creature token with flying.
			0: Draw a card. If you control three or more artifacts, draw two cards instead.
			−9: You get an emblem with "At the beginning of your end step, search your library for a permanent card, put it onto the battlefield, then shuffle."
		`,
		},
	}
}

var tezzeretArtificeMasterToken = newTezzeretArtificeMasterToken()

func newTezzeretArtificeMasterToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Thopter",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Thopter},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
