package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ManrikiGusari is the card definition for Manriki-Gusari.
//
// Type: Artifact — Equipment
// Cost: {2}
//
// Oracle text:
//
//	Equipped creature gets +1/+2 and has "{T}: Destroy target Equipment."
//	Equip {1} ({1}: Attach to target creature you control. Equip only as a sorcery.)
var ManrikiGusari = newManrikiGusari

func newManrikiGusari() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Manriki-Gusari",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +1/+2 and has \"{T}: Destroy target Equipment.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 2,
						},
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.ActivatedAbility{
									Text: "{T}: Destroy target Equipment.",
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
														Constraint: "target Equipment",
														Allow:      game.TargetAllowPermanent,
														Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Equipment}}),
													},
												},
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.Destroy{
															Object: game.TargetPermanentReference(0),
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
				game.EquipActivatedAbility(cost.Mana{cost.O(1)}),
			},
			OracleText: `
			Equipped creature gets +1/+2 and has "{T}: Destroy target Equipment."
			Equip {1} ({1}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}
