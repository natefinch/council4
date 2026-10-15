package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GarrukUnleashed is the card definition for Garruk, Unleashed.
//
// Type: Legendary Planeswalker — Garruk
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	+1: Up to one target creature gets +3/+3 and gains trample until end of turn.
//	−2: Create a 3/3 green Beast creature token. Then if an opponent controls more creatures than you, put a loyalty counter on Garruk.
//	−7: You get an emblem with "At the beginning of your end step, you may search your library for a creature card, put it onto the battlefield, then shuffle."
var GarrukUnleashed = newGarrukUnleashed

func newGarrukUnleashed() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Garruk, Unleashed",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Garruk},
			Loyalty:    opt.Val(4),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					Text:        "+1: Up to one target creature gets +3/+3 and gains trample until end of turn.",
					LoyaltyCost: 1,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MaxTargets: 1,
										Constraint: "up to one target creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.TargetPermanentReference(0)),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:       game.DurationPermanent,
													Layer:          game.LayerPowerToughnessModify,
													PowerDelta:     3,
													ToughnessDelta: 3,
												},
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.Trample},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
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
					Text:        "−2: Create a 3/3 green Beast creature token. Then if an opponent controls more creatures than you, put a loyalty counter on Garruk.",
					LoyaltyCost: -2,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(garrukUnleashedToken),
										},
									},
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.Loyalty,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text: "if an opponent controls more creatures than you",
												ControlComparison: opt.Val(game.ControlCountComparison{
													Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
													Left:      game.ControlPlayerAnyOpponent,
													Op:        compare.GreaterThan,
												}),
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
				game.LoyaltyAbility{
					Text:        "−7: You get an emblem with \"At the beginning of your end step, you may search your library for a creature card, put it onto the battlefield, then shuffle.\"",
					LoyaltyCost: -7,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateEmblem{
											EmblemAbilities: []game.Ability{
												&game.TriggeredAbility{
													Text: "At the beginning of your end step, you may search your library for a creature card, put it onto the battlefield, then shuffle.",
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
																				Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
																			},
																			Amount: game.Fixed(1),
																		},
																		Optional: true,
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
			+1: Up to one target creature gets +3/+3 and gains trample until end of turn.
			−2: Create a 3/3 green Beast creature token. Then if an opponent controls more creatures than you, put a loyalty counter on Garruk.
			−7: You get an emblem with "At the beginning of your end step, you may search your library for a creature card, put it onto the battlefield, then shuffle."
		`,
		},
	}
}

var garrukUnleashedToken = newGarrukUnleashedToken()

func newGarrukUnleashedToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Beast",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
		},
	}
}
