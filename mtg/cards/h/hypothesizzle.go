package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Hypothesizzle is the card definition for Hypothesizzle.
//
// Type: Instant
// Cost: {3}{U}{R}
//
// Oracle text:
//
//	Draw two cards. Then you may discard a nonland card. When you do, Hypothesizzle deals 4 damage to target creature.
var Hypothesizzle = newHypothesizzle

func newHypothesizzle() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Hypothesizzle",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.R,
			}),
			Colors: []color.Color{color.Red, color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.ChooseDiscardFromHand{
							Player:    game.ControllerReference(),
							Selection: game.Selection{ExcludedTypes: []types.Card{types.Land}},
						},
						Optional:      true,
						PublishResult: game.ResultKey("if-you-do"),
						LocalProducts: game.LocalProducts{
							Results: []game.ResultKey{game.ResultKey("if-you-do")},
						},
					},
					{
						Primitive: game.CreateReflexiveTrigger{
							Trigger: game.ReflexiveTriggerDef{
								Content: game.AbilityContent{
									Modes: []game.Mode{
										game.Mode{
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
												game.Instruction{
													Primitive: game.Damage{
														Amount:    game.Fixed(4),
														Recipient: game.AnyTargetDamageRecipient(0),
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
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:       "if-you-do",
							Succeeded: game.TriTrue,
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Draw two cards. Then you may discard a nonland card. When you do, Hypothesizzle deals 4 damage to target creature.
		`,
		},
	}
}
