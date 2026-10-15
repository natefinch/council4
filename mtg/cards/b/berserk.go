package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Berserk is the card definition for Berserk.
//
// Type: Instant
// Cost: {G}
//
// Oracle text:
//
//	Cast this spell only before the combat damage step.
//	Target creature gains trample and gets +X/+0 until end of turn, where X is its power. At the beginning of the next end step, destroy that creature if it attacked this turn.
var Berserk = newBerserk

func newBerserk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Berserk",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text:                           "Cast this spell only before the combat damage step.",
					CastOnlyBeforeCombatDamageStep: true,
				},
			},
			SpellAbility: opt.Val(game.Mode{
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
									Duration:    game.DurationPermanent,
									Layer:       game.LayerAbility,
									AddKeywords: []game.Keyword{game.Trample},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.ModifyPT{
							Object: game.TargetPermanentReference(0),
							PowerDelta: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountObjectPower,
								Multiplier: 1,
								Object:     game.TargetPermanentReference(0),
							}),
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								Timing: game.DelayedAtBeginningOfNextEndStep,
								Content: game.AbilityContent{
									Modes: []game.Mode{
										game.Mode{
											Sequence: []game.Instruction{
												game.Instruction{
													Primitive: game.Destroy{
														Object: game.CapturedObjectReference(),
													},
													Condition: opt.Val(game.EffectCondition{
														Condition: opt.Val(game.Condition{
															Text:                   "if it attacked this turn",
															Object:                 opt.Val(game.CapturedObjectReference()),
															ObjectAttackedThisTurn: true,
															SourceCounterKind:      counter.PlusOnePlusOne,
														}),
													}),
												},
											},
										},
									},
									MinModes: 1,
									MaxModes: 1,
								},
								CapturedObject: opt.Val(game.TargetPermanentReference(0)),
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Cast this spell only before the combat damage step.
			Target creature gains trample and gets +X/+0 until end of turn, where X is its power. At the beginning of the next end step, destroy that creature if it attacked this turn.
		`,
		},
	}
}
