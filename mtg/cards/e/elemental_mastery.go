package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ElementalMastery is the card definition for Elemental Mastery.
//
// Type: Enchantment — Aura
// Cost: {3}{R}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature has "{T}: Create X 1/1 red Elemental creature tokens with haste, where X is this creature's power. Exile them at the beginning of the next end step."
var ElementalMastery = newElementalMastery

func newElementalMastery() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Elemental Mastery",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
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
					Text: "Enchanted creature has \"{T}: Create X 1/1 red Elemental creature tokens with haste, where X is this creature's power. Exile them at the beginning of the next end step.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.ActivatedAbility{
									Text: "{T}: Create X 1/1 red Elemental creature tokens with haste, where X is this creature's power. Exile them at the beginning of the next end step.",
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
															Amount: game.Dynamic(game.DynamicAmount{
																Kind:       game.DynamicAmountObjectPower,
																Multiplier: 1,
																Object:     game.SourcePermanentReference(),
															}),
															Source:        game.TokenDef(elementalMasteryToken),
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
			Enchanted creature has "{T}: Create X 1/1 red Elemental creature tokens with haste, where X is this creature's power. Exile them at the beginning of the next end step."
		`,
		},
	}
}

var elementalMasteryToken = newElementalMasteryToken()

func newElementalMasteryToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Elemental",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
		},
	}
}
