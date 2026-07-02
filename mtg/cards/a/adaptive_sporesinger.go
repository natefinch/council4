package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AdaptiveSporesinger is the card definition for Adaptive Sporesinger.
//
// Type: Creature — Phyrexian Druid
// Cost: {2}{G}
//
// Oracle text:
//
//	Vigilance
//	When this creature enters, choose one —
//	• Target creature gets +2/+2 and gains vigilance until end of turn.
//	• Proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)
var AdaptiveSporesinger = newAdaptiveSporesinger()

func newAdaptiveSporesinger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Adaptive Sporesinger",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Druid},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
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
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Target creature gets +2/+2 and gains vigilance until end of turn.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.TargetPermanentReference(0)),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Layer:          game.LayerPowerToughnessModify,
													PowerDelta:     2,
													ToughnessDelta: 2,
												},
												game.ContinuousEffect{
													Layer: game.LayerAbility,
													AddKeywords: []game.Keyword{
														game.Vigilance,
													},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
								},
							},
							game.Mode{
								Text: "Proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)",
								Sequence: []game.Instruction{
									{
										Primitive: game.Proliferate{
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
			OracleText: `
			Vigilance
			When this creature enters, choose one —
			• Target creature gets +2/+2 and gains vigilance until end of turn.
			• Proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)
		`,
		},
	}
}
