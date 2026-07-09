package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OdricLunarchMarshal is the card definition for Odric, Lunarch Marshal.
//
// Type: Legendary Creature — Human Soldier
// Cost: {3}{W}
//
// Oracle text:
//
//	At the beginning of each combat, creatures you control gain first strike until end of turn if a creature you control has first strike. The same is true for flying, deathtouch, double strike, haste, hexproof, indestructible, lifelink, menace, reach, skulk, trample, and vigilance.
var OdricLunarchMarshal = newOdricLunarchMarshal

func newOdricLunarchMarshal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Odric, Lunarch Marshal",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Soldier},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event: game.EventBeginningOfStep,
							Step:  game.StepBeginningOfCombat,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											AddKeywords: []game.Keyword{
												game.FirstStrike,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControlsMatching: opt.Val(game.SelectionCount{
											Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.FirstStrike},
										}),
									}),
								}),
							},
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											AddKeywords: []game.Keyword{
												game.Flying,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControlsMatching: opt.Val(game.SelectionCount{
											Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.Flying},
										}),
									}),
								}),
							},
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											AddKeywords: []game.Keyword{
												game.Deathtouch,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControlsMatching: opt.Val(game.SelectionCount{
											Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.Deathtouch},
										}),
									}),
								}),
							},
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											AddKeywords: []game.Keyword{
												game.DoubleStrike,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControlsMatching: opt.Val(game.SelectionCount{
											Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.DoubleStrike},
										}),
									}),
								}),
							},
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											AddKeywords: []game.Keyword{
												game.Haste,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControlsMatching: opt.Val(game.SelectionCount{
											Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.Haste},
										}),
									}),
								}),
							},
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											AddKeywords: []game.Keyword{
												game.Hexproof,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControlsMatching: opt.Val(game.SelectionCount{
											Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.Hexproof},
										}),
									}),
								}),
							},
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											AddKeywords: []game.Keyword{
												game.Indestructible,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControlsMatching: opt.Val(game.SelectionCount{
											Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.Indestructible},
										}),
									}),
								}),
							},
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											AddKeywords: []game.Keyword{
												game.Lifelink,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControlsMatching: opt.Val(game.SelectionCount{
											Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.Lifelink},
										}),
									}),
								}),
							},
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											AddKeywords: []game.Keyword{
												game.Menace,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControlsMatching: opt.Val(game.SelectionCount{
											Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.Menace},
										}),
									}),
								}),
							},
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											AddKeywords: []game.Keyword{
												game.Reach,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControlsMatching: opt.Val(game.SelectionCount{
											Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.Reach},
										}),
									}),
								}),
							},
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											AddKeywords: []game.Keyword{
												game.Skulk,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControlsMatching: opt.Val(game.SelectionCount{
											Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.Skulk},
										}),
									}),
								}),
							},
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											AddKeywords: []game.Keyword{
												game.Trample,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControlsMatching: opt.Val(game.SelectionCount{
											Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.Trample},
										}),
									}),
								}),
							},
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											AddKeywords: []game.Keyword{
												game.Vigilance,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControlsMatching: opt.Val(game.SelectionCount{
											Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.Vigilance},
										}),
									}),
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of each combat, creatures you control gain first strike until end of turn if a creature you control has first strike. The same is true for flying, deathtouch, double strike, haste, hexproof, indestructible, lifelink, menace, reach, skulk, trample, and vigilance.
		`,
		},
	}
}
