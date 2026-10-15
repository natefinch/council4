package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RediscoverTheWay is the card definition for Rediscover the Way.
//
// Type: Enchantment — Saga
// Cost: {U}{R}{W}
//
// Oracle text:
//
//	(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
//	I, II — Look at the top three cards of your library. Put one of them into your hand and the rest on the bottom of your library in any order.
//	III — Whenever you cast a noncreature spell this turn, target creature you control gains double strike until end of turn.
var RediscoverTheWay = newRediscoverTheWay

func newRediscoverTheWay() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Rediscover the Way",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.R,
				cost.W,
			}),
			Colors:   []color.Color{color.Red, color.Blue, color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Saga},
			ChapterAbilities: []game.ChapterAbility{
				game.ChapterAbility{
					Text:     "I, II — Look at the top three cards of your library. Put one of them into your hand and the rest on the bottom of your library in any order.",
					Chapters: []int{1, 2},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Dig{
											Player:    game.ControllerReference(),
											Look:      game.Fixed(3),
											Take:      game.Fixed(1),
											Remainder: game.DigRemainderLibraryBottom,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.ChapterAbility{
					Text:     "III — Whenever you cast a noncreature spell this turn, target creature you control gains double strike until end of turn.",
					Chapters: []int{3},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateDelayedTrigger{
											Trigger: game.DelayedTriggerDef{
												Content: game.AbilityContent{
													Modes: []game.Mode{
														game.Mode{
															Targets: []game.TargetSpec{
																game.TargetSpec{
																	MinTargets: 1,
																	MaxTargets: 1,
																	Constraint: "target creature you control",
																	Allow:      game.TargetAllowPermanent,
																	Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
																},
															},
															Sequence: []game.Instruction{
																game.Instruction{
																	Primitive: game.ApplyContinuous{
																		Object: opt.Val(game.TargetPermanentReference(0)),
																		ContinuousEffects: []game.ContinuousEffect{
																			game.ContinuousEffect{
																				Duration:    game.DurationPermanent,
																				Layer:       game.LayerAbility,
																				AddKeywords: []game.Keyword{game.DoubleStrike},
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
												EventPattern: opt.Val(game.TriggerPattern{
													Event:           game.EventSpellCast,
													Controller:      game.TriggerControllerYou,
													CardSelection:   game.Selection{ExcludedTypes: []types.Card{types.Creature}},
													StackObjectKind: game.StackSpell,
													CounterKind:     counter.PlusOnePlusOne,
												}),
												Window: game.DelayedWindowThisTurn,
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
			(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
			I, II — Look at the top three cards of your library. Put one of them into your hand and the rest on the bottom of your library in any order.
			III — Whenever you cast a noncreature spell this turn, target creature you control gains double strike until end of turn.
		`,
		},
	}
}
