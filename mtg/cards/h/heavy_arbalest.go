package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// HeavyArbalest is the card definition for Heavy Arbalest.
//
// Type: Artifact — Equipment
// Cost: {3}
//
// Oracle text:
//
//	Equipped creature doesn't untap during its controller's untap step.
//	Equipped creature has "{T}: This creature deals 2 damage to any target."
//	Equip {4}
var HeavyArbalest = newHeavyArbalest

func newHeavyArbalest() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Heavy Arbalest",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature doesn't untap during its controller's untap step.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectDoesntUntap,
							AffectedAttached: true,
							Duration:         game.DurationPermanent,
						},
					},
				},
				game.StaticAbility{
					Text: "Equipped creature has \"{T}: This creature deals 2 damage to any target.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.ActivatedAbility{
									Text: "{T}: This creature deals 2 damage to any target.",
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
															Amount:       game.Fixed(2),
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
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(4)}),
			},
			OracleText: `
			Equipped creature doesn't untap during its controller's untap step.
			Equipped creature has "{T}: This creature deals 2 damage to any target."
			Equip {4}
		`,
		},
	}
}
