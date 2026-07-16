package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpringheartNantuko is the card definition for Springheart Nantuko.
//
// Type: Enchantment Creature — Insect Monk
// Cost: {1}{G}
//
// Oracle text:
//
//	Bestow {1}{G}
//	Enchanted creature gets +1/+1.
//	Landfall — Whenever a land you control enters, you may pay {1}{G} if this permanent is attached to a creature you control. If you do, create a token that's a copy of that creature. If you didn't create a token this way, create a 1/1 green Insect creature token.
var SpringheartNantuko = newSpringheartNantuko

func newSpringheartNantuko() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Springheart Nantuko",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Insect, types.Monk},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.BestowStaticAbility(cost.Mana{cost.O(1), cost.G}, &game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Land}},
						},
					},
					Content: game.Mode{
						Text: "Landfall — Whenever a land you control enters, you may pay {1}{G} if this permanent is attached to a creature you control. If you do, create a token that's a copy of that creature. If you didn't create a token this way, create a 1/1 green Insect creature token.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Pay {1}{G}?",
										ManaCost: opt.Val(cost.Mana{
											cost.O(1),
											cost.G,
										}),
									},
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										Object:        opt.Val(game.SourceAttachedPermanentReference()),
										ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									}),
								}),
								PublishResult: game.ResultKey("springheart-landfall-paid"),
							},
							{
								Primitive: game.CreateReflexiveTrigger{
									Trigger: game.ReflexiveTriggerDef{
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.CreateToken{
														Amount: game.Fixed(1),
														Source: game.TokenCopyOf(game.TokenCopySpec{
															Source: game.TokenCopySourceObject,
															Object: game.SourceAttachedPermanentReference(),
														}),
													},
													PublishResult: game.ResultKey("springheart-landfall-copied"),
												},
												{
													Primitive: game.CreateToken{
														Amount: game.Fixed(1),
														Source: game.TokenDef(springheartNantukoToken),
													},
													ResultGate: opt.Val(game.InstructionResultGate{
														Key:       "springheart-landfall-copied",
														Succeeded: game.TriFalse,
													}),
												},
											},
										}.Ability(),
									},
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "springheart-landfall-paid",
									Succeeded: game.TriTrue,
								}),
							},
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(springheartNantukoToken),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "springheart-landfall-paid",
									Succeeded: game.TriFalse,
								}),
							},
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(springheartNantukoToken),
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										Negate:        true,
										Object:        opt.Val(game.SourceAttachedPermanentReference()),
										ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									}),
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Bestow {1}{G}
			Enchanted creature gets +1/+1.
			Landfall — Whenever a land you control enters, you may pay {1}{G} if this permanent is attached to a creature you control. If you do, create a token that's a copy of that creature. If you didn't create a token this way, create a 1/1 green Insect creature token.
		`,
		},
	}
}

var springheartNantukoToken = newSpringheartNantukoToken()

func newSpringheartNantukoToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Insect",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Insect},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
