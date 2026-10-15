package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// InvocationOfSaintTraft is the card definition for Invocation of Saint Traft.
//
// Type: Enchantment — Aura
// Cost: {1}{W}{U}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature has "Whenever this creature attacks, create a 4/4 white Angel creature token with flying that's tapped and attacking. Exile that token at end of combat."
var InvocationOfSaintTraft = newInvocationOfSaintTraft

func newInvocationOfSaintTraft() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Invocation of Saint Traft",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.U,
			}),
			Colors:   []color.Color{color.Blue, color.White},
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
					Text: "Enchanted creature has \"Whenever this creature attacks, create a 4/4 white Angel creature token with flying that's tapped and attacking. Exile that token at end of combat.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.TriggeredAbility{
									Text: "Whenever this creature attacks, create a 4/4 white Angel creature token with flying that's tapped and attacking. Exile that token at end of combat.",
									Trigger: game.TriggerCondition{
										Type: game.TriggerWhenever,
										Pattern: game.TriggerPattern{
											Event:           game.EventAttackerDeclared,
											Source:          game.TriggerSourceSelf,
											StackObjectKind: game.StackSpell,
											CounterKind:     counter.PlusOnePlusOne,
										},
									},
									Content: game.AbilityContent{
										Modes: []game.Mode{
											game.Mode{
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.CreateToken{
															Amount:         game.Fixed(1),
															Source:         game.TokenDef(invocationOfSaintTraftToken),
															EntryTapped:    true,
															EntryAttacking: true,
															PublishLinked:  game.LinkedKey("sequence-effect-0-product"),
														},
														LocalProducts: game.LocalProducts{
															Links: []game.LinkedKey{game.LinkedKey("sequence-effect-0-product")},
														},
													},
													game.Instruction{
														Primitive: game.CreateDelayedTrigger{
															Trigger: game.DelayedTriggerDef{
																Timing: game.DelayedAtEndOfCombat,
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
			Enchanted creature has "Whenever this creature attacks, create a 4/4 white Angel creature token with flying that's tapped and attacking. Exile that token at end of combat."
		`,
		},
	}
}

var invocationOfSaintTraftToken = newInvocationOfSaintTraftToken()

func newInvocationOfSaintTraftToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Angel",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
