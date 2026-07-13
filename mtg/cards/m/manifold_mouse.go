package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ManifoldMouse is the card definition for ManifoldMouse.
//
// Type: Creature — Mouse Soldier
// Cost: {1}{R}
//
// Oracle text:
//
//	Offspring {2} (You may pay an additional {2} as you cast this spell. If you do, when this creature enters, create a 1/1 token copy of it.)
//	At the beginning of combat on your turn, target Mouse you control gains your choice of double strike or trample until end of turn.
var ManifoldMouse = newManifoldMouse

func newManifoldMouse() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Manifold Mouse",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Mouse, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.OffspringStaticAbility(cost.Mana{cost.O(2)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.OffspringEnterTriggeredAbility(),
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepBeginningOfCombat,
						},
					},
					Content: game.AbilityContent{
						SharedTargets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target Mouse you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Mouse")}, Controller: game.ControllerYou}),
							},
						},
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.TargetPermanentReference(0)),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Layer: game.LayerAbility,
													AddKeywords: []game.Keyword{
														game.DoubleStrike,
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
											Object: opt.Val(game.TargetPermanentReference(0)),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Layer: game.LayerAbility,
													AddKeywords: []game.Keyword{
														game.Trample,
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
			},
			OracleText: `
			Offspring {2} (You may pay an additional {2} as you cast this spell. If you do, when this creature enters, create a 1/1 token copy of it.)
			At the beginning of combat on your turn, target Mouse you control gains your choice of double strike or trample until end of turn.
		`,
		},
	}
}
