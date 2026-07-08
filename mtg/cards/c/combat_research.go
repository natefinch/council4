package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CombatResearch is the card definition for Combat Research.
//
// Type: Enchantment — Aura
// Cost: {U}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature has "Whenever this creature deals combat damage to a player, draw a card."
//	As long as enchanted creature is legendary, it gets +1/+1 and has ward {1}. (Whenever enchanted creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {1}.)
var CombatResearch = newCombatResearch

func newCombatResearch() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Combat Research",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
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
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.TriggeredAbility{
									Trigger: game.TriggerCondition{
										Type: game.TriggerWhenever,
										Pattern: game.TriggerPattern{
											Event:               game.EventDamageDealt,
											Source:              game.TriggerSourceSelf,
											Subject:             game.TriggerSubjectDamageSource,
											RequireCombatDamage: true,
											DamageRecipient:     game.DamageRecipientPlayer,
										},
									},
									Content: game.Mode{
										Sequence: []game.Instruction{
											{
												Primitive: game.Draw{
													Amount: game.Fixed(1),
													Player: game.ControllerReference(),
												},
											},
										},
									}.Ability(),
								}),
							},
						},
					},
				},
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Object:        opt.Val(game.SourceAttachedPermanentReference()),
						ObjectMatches: opt.Val(game.Selection{Supertypes: []types.Super{types.Legendary}}),
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.WardStaticAbility(cost.Mana{cost.O(1)})),
							},
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature has "Whenever this creature deals combat damage to a player, draw a card."
			As long as enchanted creature is legendary, it gets +1/+1 and has ward {1}. (Whenever enchanted creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {1}.)
		`,
		},
	}
}
