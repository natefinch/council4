package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TheOzolith is the card definition for The Ozolith.
//
// Type: Legendary Artifact
// Cost: {1}
//
// Oracle text:
//
//	Whenever a creature you control leaves the battlefield, if it had counters on it, put those counters on The Ozolith.
//	At the beginning of combat on your turn, if The Ozolith has counters on it, you may move all counters from The Ozolith onto target creature.
var TheOzolith = newTheOzolith

func newTheOzolith() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "The Ozolith",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever a creature you control leaves the battlefield, if it had counters on it, put those counters on The Ozolith.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventZoneChanged,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							MatchFromZone:    true,
							FromZone:         zone.Battlefield,
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
						InterveningIf:                          "if it had counters on it",
						InterveningIfEventPermanentHadCounters: true,
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCounters{
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
											Source: game.CounterSourceSpec{
												Kind: game.CounterSourceEventPermanent,
											},
											AllKinds: true,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.TriggeredAbility{
					Text: "At the beginning of combat on your turn, if The Ozolith has counters on it, you may move all counters from The Ozolith onto target creature.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepBeginningOfCombat,
						},
						InterveningIf: "if The Ozolith has counters on it",
						InterveningCondition: opt.Val(game.Condition{
							Text:              "if The Ozolith has counters on it",
							Object:            opt.Val(game.SourcePermanentReference()),
							ObjectMatches:     opt.Val(game.Selection{MatchAnyCounter: true}),
							SourceCounterKind: counter.PlusOnePlusOne,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
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
									game.Instruction{
										Primitive: game.MoveCounters{
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.PlusOnePlusOne,
											Source: game.CounterSourceSpec{
												Kind: game.CounterSourceSelf,
											},
											AllKinds: true,
										},
										Optional: true,
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Whenever a creature you control leaves the battlefield, if it had counters on it, put those counters on The Ozolith.
			At the beginning of combat on your turn, if The Ozolith has counters on it, you may move all counters from The Ozolith onto target creature.
		`,
		},
	}
}
