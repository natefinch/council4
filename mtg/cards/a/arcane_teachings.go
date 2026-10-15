package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ArcaneTeachings is the card definition for Arcane Teachings.
//
// Type: Enchantment — Aura
// Cost: {2}{R}
//
// Oracle text:
//
//	Enchant creature (Target a creature as you cast this. This card enters attached to that creature.)
//	Enchanted creature gets +2/+2 and has "{T}: This creature deals 1 damage to any target."
var ArcaneTeachings = newArcaneTeachings

func newArcaneTeachings() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Arcane Teachings",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
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
					Text: "Enchanted creature gets +2/+2 and has \"{T}: This creature deals 1 damage to any target.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.ActivatedAbility{
									Text: "{T}: This creature deals 1 damage to any target.",
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
												Targets: []game.TargetSpec{
													game.TargetSpec{
														MinTargets: 1,
														MaxTargets: 1,
														Constraint: "any target",
														Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
														Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
													},
												},
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.Damage{
															Amount:       game.Fixed(1),
															Recipient:    game.AnyTargetDamageRecipient(0),
															DamageSource: opt.Val(game.SourcePermanentReference()),
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
			Enchant creature (Target a creature as you cast this. This card enters attached to that creature.)
			Enchanted creature gets +2/+2 and has "{T}: This creature deals 1 damage to any target."
		`,
		},
	}
}
