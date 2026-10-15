package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EnsouledScimitar is the card definition for Ensouled Scimitar.
//
// Type: Artifact — Equipment
// Cost: {3}
//
// Oracle text:
//
//	{3}: This artifact becomes a 1/5 Spirit artifact creature with flying until end of turn. (Equipment that's a creature can't equip a creature.)
//	Equipped creature gets +1/+5.
//	Equip {2} ({2}: Attach to target creature you control. Equip only as a sorcery.)
var EnsouledScimitar = newEnsouledScimitar

func newEnsouledScimitar() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Ensouled Scimitar",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +1/+5.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 5,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{3}: This artifact becomes a 1/5 Spirit artifact creature with flying until end of turn. (Equipment that's a creature can't equip a creature.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.SourcePermanentReference()),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerType,
													AddTypes:    []types.Card{types.Creature, types.Artifact},
													AddSubtypes: []types.Sub{types.Spirit},
												},
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.Flying},
												},
												game.ContinuousEffect{
													Duration: game.DurationPermanent,
													Layer:    game.LayerPowerToughnessSet,
													SetPower: opt.Val(game.PT{
														Value: 1,
													}),
													SetToughness: opt.Val(game.PT{
														Value: 5,
													}),
												},
											},
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
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			{3}: This artifact becomes a 1/5 Spirit artifact creature with flying until end of turn. (Equipment that's a creature can't equip a creature.)
			Equipped creature gets +1/+5.
			Equip {2} ({2}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}
