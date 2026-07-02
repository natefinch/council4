package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UmezawaSJitte is the card definition for Umezawa's Jitte.
//
// Type: Legendary Artifact — Equipment
// Cost: {2}
//
// Oracle text:
//
//	Whenever equipped creature deals combat damage, put two charge counters on Umezawa's Jitte.
//	Remove a charge counter from Umezawa's Jitte: Choose one —
//	• Equipped creature gets +2/+2 until end of turn.
//	• Target creature gets -1/-1 until end of turn.
//	• You gain 2 life.
//	Equip {2}
var UmezawaSJitte = newUmezawaSJitte()

func newUmezawaSJitte() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Umezawa's Jitte",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Equipment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: `Remove a charge counter from Umezawa's Jitte: Choose one —
		• Equipped creature gets +2/+2 until end of turn.
		• Target creature gets -1/-1 until end of turn.
		• You gain 2 life.`,
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a charge counter from Umezawa's Jitte",
							Amount:      1,
							CounterKind: counter.Charge,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Equipped creature gets +2/+2 until end of turn.",
								Sequence: []game.Instruction{
									{
										Primitive: game.ApplyContinuous{
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Layer:          game.LayerPowerToughnessModify,
													Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
													PowerDelta:     2,
													ToughnessDelta: 2,
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
								},
							},
							game.Mode{
								Text: "Target creature gets -1/-1 until end of turn.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.ModifyPT{
											Object:         game.TargetPermanentReference(0),
											PowerDelta:     game.Fixed(-1),
											ToughnessDelta: game.Fixed(-1),
											Duration:       game.DurationUntilEndOfTurn,
										},
									},
								},
							},
							game.Mode{
								Text: "You gain 2 life.",
								Sequence: []game.Instruction{
									{
										Primitive: game.GainLife{
											Amount: game.Fixed(2),
											Player: game.ControllerReference(),
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                 game.EventDamageDealt,
							Source:                game.TriggerSourceAttachedPermanent,
							Subject:               game.TriggerSubjectDamageSource,
							RequireCombatDamage:   true,
							DamageSourceSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(2),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Charge,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever equipped creature deals combat damage, put two charge counters on Umezawa's Jitte.
			Remove a charge counter from Umezawa's Jitte: Choose one —
			• Equipped creature gets +2/+2 until end of turn.
			• Target creature gets -1/-1 until end of turn.
			• You gain 2 life.
			Equip {2}
		`,
		},
	}
}
