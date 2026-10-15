package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ObNixilisOfTheBlackOath is the card definition for Ob Nixilis of the Black Oath.
//
// Type: Legendary Planeswalker — Nixilis
// Cost: {3}{B}{B}
//
// Oracle text:
//
//	+2: Each opponent loses 1 life. You gain life equal to the life lost this way.
//	−2: Create a 5/5 black Demon creature token with flying. You lose 2 life.
//	−8: You get an emblem with "{1}{B}, Sacrifice a creature: You gain X life and draw X cards, where X is the sacrificed creature's power."
//	Ob Nixilis of the Black Oath can be your commander.
var ObNixilisOfTheBlackOath = newObNixilisOfTheBlackOath

func newObNixilisOfTheBlackOath() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Ob Nixilis of the Black Oath",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
			}),
			Colors:         []color.Color{color.Black},
			CanBeCommander: true,
			Supertypes:     []types.Super{types.Legendary},
			Types:          []types.Card{types.Planeswalker},
			Subtypes:       []types.Sub{types.Nixilis},
			Loyalty:        opt.Val(3),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					Text:        "+2: Each opponent loses 1 life. You gain life equal to the life lost this way.",
					LoyaltyCost: 2,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.LoseLife{
											Amount:      game.Fixed(1),
											PlayerGroup: game.OpponentsReference(),
										},
										PublishResult: game.ResultKey("life-change"),
									},
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:      game.DynamicAmountPreviousEffectResult,
												ResultKey: game.ResultKey("life-change"),
											}),
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
					Text:        "−2: Create a 5/5 black Demon creature token with flying. You lose 2 life.",
					LoyaltyCost: -2,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(obNixilisOfTheBlackOathToken),
										},
									},
									game.Instruction{
										Primitive: game.LoseLife{
											Amount: game.Fixed(2),
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
					Text:        "−8: You get an emblem with \"{1}{B}, Sacrifice a creature: You gain X life and draw X cards, where X is the sacrificed creature's power.\"",
					LoyaltyCost: -8,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateEmblem{
											EmblemAbilities: []game.Ability{
												&game.ActivatedAbility{
													Text: "{1}{B}, Sacrifice a creature: You gain X life and draw X cards, where X is the sacrificed creature's power.",
													ManaCost: opt.Val(cost.Mana{
														cost.O(1),
														cost.B,
													}),
													AdditionalCosts: []cost.Additional{
														cost.Additional{
															Kind:               cost.AdditionalSacrifice,
															Text:               "Sacrifice a creature",
															Amount:             1,
															MatchPermanentType: true,
															PermanentType:      types.Creature,
															CounterKind:        counter.PlusOnePlusOne,
														},
													},
													ZoneOfFunction: zone.Battlefield,
													Content: game.AbilityContent{
														Modes: []game.Mode{
															game.Mode{
																Sequence: []game.Instruction{
																	game.Instruction{
																		Primitive: game.GainLife{
																			Amount: game.Dynamic(game.DynamicAmount{
																				Kind: game.DynamicAmountX,
																			}),
																			Player: game.ControllerReference(),
																		},
																	},
																	game.Instruction{
																		Primitive: game.Draw{
																			Amount: game.Dynamic(game.DynamicAmount{
																				Kind:       game.DynamicAmountObjectPower,
																				Multiplier: 1,
																				Object:     game.SacrificedCostReference(),
																			}),
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
			+2: Each opponent loses 1 life. You gain life equal to the life lost this way.
			−2: Create a 5/5 black Demon creature token with flying. You lose 2 life.
			−8: You get an emblem with "{1}{B}, Sacrifice a creature: You gain X life and draw X cards, where X is the sacrificed creature's power."
			Ob Nixilis of the Black Oath can be your commander.
		`,
		},
	}
}

var obNixilisOfTheBlackOathToken = newObNixilisOfTheBlackOathToken()

func newObNixilisOfTheBlackOathToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Demon",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Demon},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
