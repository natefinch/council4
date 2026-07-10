package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BlasterCombatDJ is the card definition for Blaster, Combat DJ // Blaster, Morale Booster.
//
// Type: Legendary Artifact Creature — Robot // Legendary Artifact
// Face: Blaster, Morale Booster — Legendary Artifact
//
// Oracle text:
//
//	More Than Meets the Eye {1}{R}{G} (You may cast this card converted for {1}{R}{G}.)
//	Other nontoken artifact creatures and Vehicles you control have modular 1. (They enter with an additional +1/+1 counter on them. When they die, you may put their +1/+1 counters on target artifact creature.)
//	Whenever you put one or more +1/+1 counters on Blaster, convert it.
var BlasterCombatDJ = newBlasterCombatDJ

func newBlasterCombatDJ() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Blaster, Combat DJ",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.G,
			}),
			Colors:     []color.Color{color.Green, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact, types.Creature},
			Subtypes:   []types.Sub{types.Robot},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.BattlefieldGroup(game.Selection{AnyOf: []game.Selection{game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}, NonToken: true}, game.Selection{SubtypesAny: []types.Sub{types.Sub("Vehicle")}, NonToken: true}}, Controller: game.ControllerYou, ExcludeSource: true}),
							AddAbilities: []game.Ability{
								new(game.TriggeredAbility{
									Trigger: game.TriggerCondition{
										Type: game.TriggerWhen,
										Pattern: game.TriggerPattern{
											Event:            game.EventPermanentDied,
											Source:           game.TriggerSourceSelf,
											SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
										},
									},
									Optional: true,
									Content: game.Mode{
										Targets: []game.TargetSpec{
											game.TargetSpec{
												MinTargets: 1,
												MaxTargets: 1,
												Constraint: "target artifact creature",
												Allow:      game.TargetAllowPermanent,
												Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}}),
											},
										},
										Sequence: []game.Instruction{
											{
												Primitive: game.MoveCounters{
													Amount: game.Dynamic(game.DynamicAmount{
														Kind:        game.DynamicAmountObjectCounters,
														CounterKind: counter.PlusOnePlusOne,
														Object:      game.SourcePermanentReference(),
													}),
													Object:      game.TargetPermanentReference(0),
													CounterKind: counter.PlusOnePlusOne,
													Source:      game.CounterSourceSpec{Kind: game.CounterSourceSelf},
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventCountersAdded,
							Source:           game.TriggerSourceSelf,
							CauseController:  game.TriggerControllerYou,
							OneOrMore:        true,
							MatchCounterKind: true,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Transform{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersGroupReplacement("Each other nontoken artifact creatures or nontoken Vehicles you control enters with an additional +1/+1 counter on it.", &game.Selection{AnyOf: []game.Selection{game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}, NonToken: true}, game.Selection{SubtypesAny: []types.Sub{types.Sub("Vehicle")}, NonToken: true}}, Controller: game.ControllerYou, ExcludeSource: true}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "More Than Meets the Eye",
					ManaCost: opt.Val(cost.Mana{cost.O(1), cost.R, cost.G}),
					Mechanic: cost.AlternativeMechanicMoreThanMeetsTheEye,
				},
			},
			OracleText: `
			More Than Meets the Eye {1}{R}{G} (You may cast this card converted for {1}{R}{G}.)
			Other nontoken artifact creatures and Vehicles you control have modular 1. (They enter with an additional +1/+1 counter on them. When they die, you may put their +1/+1 counters on target artifact creature.)
			Whenever you put one or more +1/+1 counters on Blaster, convert it.
		`,
		},
		Layout: game.LayoutTransform,
		Back: opt.Val(game.CardFace{
			Name:       "Blaster, Morale Booster",
			Colors:     []color.Color{color.Green, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{X}, {T}: Move X +1/+1 counters from Blaster onto another target artifact. That artifact gains haste until end of turn. If Blaster has no +1/+1 counters on it, convert it. Activate only as a sorcery.",
					ManaCost:        opt.Val(cost.Mana{cost.X}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Timing:          game.SorceryOnly,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "another target artifact",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}, ExcludeSource: true}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCounters{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind: game.DynamicAmountX,
									}),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.PlusOnePlusOne,
									Source:      game.CounterSourceSpec{Kind: game.CounterSourceSelf},
								},
							},
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Haste,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
							{
								Primitive: game.Transform{
									Object: game.SourcePermanentReference(),
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										Negate:        true,
										Object:        opt.Val(game.SourcePermanentReference()),
										ObjectMatches: opt.Val(game.Selection{RequiredCounterCount: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 1}), RequiredCounter: counter.PlusOnePlusOne}),
									}),
								}),
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Optional: true,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target artifact creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCounters{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:        game.DynamicAmountObjectCounters,
										CounterKind: counter.PlusOnePlusOne,
										Object:      game.SourcePermanentReference(),
									}),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.PlusOnePlusOne,
									Source:      game.CounterSourceSpec{Kind: game.CounterSourceSelf},
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with three +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 3}),
			},
			OracleText: `
			Modular 3
			{X}, {T}: Move X +1/+1 counters from Blaster onto another target artifact. That artifact gains haste until end of turn. If Blaster has no +1/+1 counters on it, convert it. Activate only as a sorcery.
		`,
		}),
	}
}
