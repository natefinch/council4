package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WingnutBatOnTheBelfry is the card definition for Wingnut, Bat on the Belfry.
//
// Type: Legendary Creature — Bat Mutant
// Cost: {1}{R}
//
// Oracle text:
//
//	Alliance — Whenever another creature you control enters, Wingnut gains your choice of flying, menace, or haste until end of turn.
//	Whenever Wingnut attacks, each other attacking creature gets +1/+0 until end of turn.
var WingnutBatOnTheBelfry = newWingnutBatOnTheBelfry

func newWingnutBatOnTheBelfry() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Wingnut, Bat on the Belfry",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Bat, types.Mutant},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							ExcludeSelf:      true,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.SourceCardPermanentReference()),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Layer: game.LayerAbility,
													AddKeywords: []game.Keyword{
														game.Flying,
													},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
								},
							},
							game.Mode{
								Sequence: []game.Instruction{
									{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.SourceCardPermanentReference()),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Layer: game.LayerAbility,
													AddKeywords: []game.Keyword{
														game.Menace,
													},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
								},
							},
							game.Mode{
								Sequence: []game.Instruction{
									{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.SourceCardPermanentReference()),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Layer: game.LayerAbility,
													AddKeywords: []game.Keyword{
														game.Haste,
													},
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
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:      game.LayerPowerToughnessModify,
											Group:      game.BattlefieldGroupExcluding(game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}, game.SourcePermanentReference()),
											PowerDelta: 1,
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Alliance — Whenever another creature you control enters, Wingnut gains your choice of flying, menace, or haste until end of turn.
			Whenever Wingnut attacks, each other attacking creature gets +1/+0 until end of turn.
		`,
		},
	}
}
