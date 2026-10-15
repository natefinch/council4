package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ScreamingShield is the card definition for Screaming Shield.
//
// Type: Artifact — Equipment
// Cost: {1}
//
// Oracle text:
//
//	Equipped creature gets +0/+3 and has "{2}, {T}: Target player mills three cards."
//	Equip {3} ({3}: Attach to target creature you control. Equip only as a sorcery.)
var ScreamingShield = newScreamingShield

func newScreamingShield() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Screaming Shield",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +0/+3 and has \"{2}, {T}: Target player mills three cards.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							ToughnessDelta: 3,
						},
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.ActivatedAbility{
									Text: "{2}, {T}: Target player mills three cards.",
									ManaCost: opt.Val(cost.Mana{
										cost.O(2),
									}),
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
														Constraint: "Target player",
														Allow:      game.TargetAllowPlayer,
													},
												},
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.MoveTopOfLibrary{
															Amount:      game.Fixed(3),
															Player:      game.TargetPlayerReference(0),
															Destination: zone.Graveyard,
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
				game.EquipActivatedAbility(cost.Mana{cost.O(3)}),
			},
			OracleText: `
			Equipped creature gets +0/+3 and has "{2}, {T}: Target player mills three cards."
			Equip {3} ({3}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}
