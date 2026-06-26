package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LostJitte is the card definition for Lost Jitte.
//
// Type: Legendary Artifact — Equipment
// Cost: {1}
//
// Oracle text:
//
//	Whenever equipped creature deals combat damage, put a charge counter on Lost Jitte.
//	Remove a charge counter from Lost Jitte: Choose one —
//	• Untap target land.
//	• Target creature can't block this turn.
//	• Put a +1/+1 counter on equipped creature.
//	Equip {1}
var LostJitte = newLostJitte()

func newLostJitte() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Lost Jitte",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Equipment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: `Remove a charge counter from Lost Jitte: Choose one —
		• Untap target land.
		• Target creature can't block this turn.
		• Put a +1/+1 counter on equipped creature.`,
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a charge counter from Lost Jitte",
							Amount:      1,
							CounterKind: counter.Charge,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Untap target land.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target land",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(0),
										},
									},
								},
							},
							game.Mode{
								Text: "Target creature can't block this turn.",
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
										Primitive: game.ApplyRule{
											Object: opt.Val(game.TargetPermanentReference(0)),
											RuleEffects: []game.RuleEffect{
												game.RuleEffect{
													Kind: game.RuleEffectCantBlock,
												},
											},
											Duration: game.DurationThisTurn,
										},
									},
								},
							},
							game.Mode{
								Text: "Put a +1/+1 counter on equipped creature.",
								Sequence: []game.Instruction{
									{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourceAttachedPermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.EquipActivatedAbility(cost.Mana{cost.O(1)}),
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
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Charge,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever equipped creature deals combat damage, put a charge counter on Lost Jitte.
			Remove a charge counter from Lost Jitte: Choose one —
			• Untap target land.
			• Target creature can't block this turn.
			• Put a +1/+1 counter on equipped creature.
			Equip {1}
		`,
		},
	}
}
