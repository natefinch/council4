package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Idolized is the card definition for Idolized.
//
// Type: Enchantment — Aura
// Cost: {1}{W}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature has "Whenever this creature attacks alone, it gets +X/+X until end of turn, where X is the number of nonland permanents you control."
var Idolized = newIdolized

func newIdolized() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Idolized",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:   []color.Color{color.White},
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
					Text: "Enchanted creature has \"Whenever this creature attacks alone, it gets +X/+X until end of turn, where X is the number of nonland permanents you control.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.TriggeredAbility{
									Text: "Whenever this creature attacks alone, it gets +X/+X until end of turn, where X is the number of nonland permanents you control.",
									Trigger: game.TriggerCondition{
										Type: game.TriggerWhenever,
										Pattern: game.TriggerPattern{
											Event:           game.EventAttackerDeclared,
											Source:          game.TriggerSourceSelf,
											StackObjectKind: game.StackSpell,
											AttackAlone:     true,
											CounterKind:     counter.PlusOnePlusOne,
										},
									},
									Content: game.AbilityContent{
										Modes: []game.Mode{
											game.Mode{
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.ModifyPT{
															Object: game.EventPermanentReference(),
															PowerDelta: game.Dynamic(game.DynamicAmount{
																Kind:       game.DynamicAmountCountSelector,
																Multiplier: 1,
																Group:      game.BattlefieldGroup(game.Selection{ExcludedTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
															}),
															ToughnessDelta: game.Dynamic(game.DynamicAmount{
																Kind:       game.DynamicAmountCountSelector,
																Multiplier: 1,
																Group:      game.BattlefieldGroup(game.Selection{ExcludedTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
															}),
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
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature has "Whenever this creature attacks alone, it gets +X/+X until end of turn, where X is the number of nonland permanents you control."
		`,
		},
	}
}
