package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AvariceAmulet is the card definition for Avarice Amulet.
//
// Type: Artifact — Equipment
// Cost: {4}
//
// Oracle text:
//
//	Equipped creature gets +2/+0 and has vigilance and "At the beginning of your upkeep, draw a card."
//	Whenever equipped creature dies, target opponent gains control of this Equipment.
//	Equip {2} ({2}: Attach to target creature you control. Equip only as a sorcery.)
var AvariceAmulet = newAvariceAmulet

func newAvariceAmulet() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Avarice Amulet",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:      game.LayerPowerToughnessModify,
							Group:      game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta: 2,
						},
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.Vigilance,
							},
						},
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.TriggeredAbility{
									Trigger: game.TriggerCondition{
										Type: game.TriggerAt,
										Pattern: game.TriggerPattern{
											Event:      game.EventBeginningOfStep,
											Controller: game.TriggerControllerYou,
											Step:       game.StepUpkeep,
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
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceAttachedPermanent,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target opponent",
								Allow:      game.TargetAllowPlayer,
								Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.SourcePermanentReference()),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:            game.LayerControl,
											NewControllerRef: opt.Val(game.TargetPlayerReference(0)),
										},
									},
									Duration: game.DurationPermanent,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Equipped creature gets +2/+0 and has vigilance and "At the beginning of your upkeep, draw a card."
			Whenever equipped creature dies, target opponent gains control of this Equipment.
			Equip {2} ({2}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}
