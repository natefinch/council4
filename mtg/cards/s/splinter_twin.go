package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SplinterTwin is the card definition for Splinter Twin.
//
// Type: Enchantment — Aura
// Cost: {2}{R}{R}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature has "{T}: Create a token that's a copy of this creature, except it has haste. Exile that token at the beginning of the next end step."
var SplinterTwin = newSplinterTwin

func newSplinterTwin() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Splinter Twin",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					Text: "Enchanted creature has \"{T}: Create a token that's a copy of this creature, except it has haste. Exile that token at the beginning of the next end step.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.ActivatedAbility{
									Text: "{T}: Create a token that's a copy of this creature, except it has haste. Exile that token at the beginning of the next end step.",
									AdditionalCosts: []cost.Additional{
										cost.Additional{
											Kind:        cost.AdditionalTap,
											CounterKind: counter.PlusOnePlusOne,
										},
									},
									ZoneOfFunction: zone.Battlefield,
									Content: game.AbilityContent{
										Modes: []game.Mode{
											game.Mode{
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.CreateToken{
															Amount: game.Fixed(1),
															Source: game.TokenCopyOf(game.TokenCopySpec{
																Source:      game.TokenCopySourceObject,
																Object:      game.SourcePermanentReference(),
																AddKeywords: []game.Keyword{game.Haste},
															}),
															PublishLinked: game.LinkedKey("sequence-effect-0-product"),
														},
														LocalProducts: game.LocalProducts{
															Links: []game.LinkedKey{game.LinkedKey("sequence-effect-0-product")},
														},
													},
													game.Instruction{
														Primitive: game.CreateDelayedTrigger{
															Trigger: game.DelayedTriggerDef{
																Timing: game.DelayedAtBeginningOfNextEndStep,
																Content: game.AbilityContent{
																	Modes: []game.Mode{
																		game.Mode{
																			Sequence: []game.Instruction{
																				game.Instruction{
																					Primitive: game.MovePermanent{
																						Group:       game.CapturedObjectsGroup(),
																						Destination: zone.Exile,
																					},
																				},
																			},
																		},
																	},
																	MinModes: 1,
																	MaxModes: 1,
																},
																CapturedObjectGroup: opt.Val(game.LinkedObjectReference("sequence-effect-0-product")),
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
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature has "{T}: Create a token that's a copy of this creature, except it has haste. Exile that token at the beginning of the next end step."
		`,
		},
	}
}
