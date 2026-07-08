package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TheFireNationDrill is the card definition for The Fire Nation Drill.
//
// Type: Legendary Artifact — Vehicle
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	Trample
//	When The Fire Nation Drill enters, you may tap it. When you do, destroy target creature with power 4 or less.
//	{1}: Permanents your opponents control lose hexproof and indestructible until end of turn.
//	Crew 2
var TheFireNationDrill = newTheFireNationDrill

func newTheFireNationDrill() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "The Fire Nation Drill",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Vehicle},
			Power:      opt.Val(game.PT{Value: 6}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{1}: Permanents your opponents control lose hexproof and indestructible until end of turn.",
					ManaCost:       opt.Val(cost.Mana{cost.O(1)}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{Controller: game.ControllerOpponent}),
											RemoveKeywords: []game.Keyword{
												game.Hexproof,
												game.Indestructible,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
				game.CrewActivatedAbility(2),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Tap{
									Object: game.EventPermanentReference(),
								},
								Optional:      true,
								PublishResult: game.ResultKey("if-you-do"),
							},
							{
								Primitive: game.CreateReflexiveTrigger{
									Trigger: game.ReflexiveTriggerDef{
										Content: game.Mode{
											Targets: []game.TargetSpec{
												game.TargetSpec{
													MinTargets: 1,
													MaxTargets: 1,
													Constraint: "target creature with power 4 or less",
													Allow:      game.TargetAllowPermanent,
													Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 4})}),
												},
											},
											Sequence: []game.Instruction{
												{
													Primitive: game.Destroy{
														Object: game.TargetPermanentReference(0),
													},
												},
											},
										}.Ability(),
									},
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "if-you-do",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Trample
			When The Fire Nation Drill enters, you may tap it. When you do, destroy target creature with power 4 or less.
			{1}: Permanents your opponents control lose hexproof and indestructible until end of turn.
			Crew 2
		`,
		},
	}
}
