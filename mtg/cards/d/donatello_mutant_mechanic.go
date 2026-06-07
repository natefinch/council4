package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DonatelloMutantMechanic is the card definition for Donatello, Mutant Mechanic.
//
// Type: types.Legendary Creature — Mutant Ninja Turtle
// Cost: {3}{U}
//
// Oracle text:
//
//	{T}: Put three +1/+1 counters on target artifact you control. If it isn't a creature, it becomes a 0/0 Robot creature in addition to its other types. Activate only as a sorcery.
//	Whenever an artifact you control is put into a graveyard from the battlefield, if it had counters on it, put those counters on up to one target artifact or creature you control.
var DonatelloMutantMechanic = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name: "Donatello, Mutant Mechanic",
		ManaCost: opt.Val(cost.Mana{
			cost.O(3),
			cost.U,
		}),
		Colors:     []color.Color{color.Blue},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Mutant, types.Ninja, types.Turtle},
		Power:      opt.Val(game.PT{Value: 3}),
		Toughness:  opt.Val(game.PT{Value: 5}),
		OracleText: `
			{T}: Put three +1/+1 counters on target artifact you control. If it isn't a creature, it becomes a 0/0 Robot creature in addition to its other types. Activate only as a sorcery.
			Whenever an artifact you control is put into a graveyard from the battlefield, if it had counters on it, put those counters on up to one target artifact or creature you control.
		`,
		ActivatedAbilities: []game.ActivatedAbilityBody{
			{
				Text: `
					{T}: Put three +1/+1 counters on target artifact you control. If it isn't a creature, it becomes a 0/0 Robot creature in addition to its other types. Activate only as a sorcery.
				`,
				AdditionalCosts: cost.Tap,
				Timing:          game.SorceryOnly,
				Content: game.Mode{
					Targets: []game.TargetSpec{
						{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "artifact you control",
						},
					},
					Sequence: []game.Instruction{
						{
							Primitive: game.AddCounter{
								Amount:      game.Fixed(3),
								TargetIndex: 0,
								CounterKind: counter.PlusOnePlusOne,
							},
						},
						{
							Primitive: game.ApplyContinuous{
								TargetIndex: 0,
								ContinuousEffects: []game.ContinuousEffect{
									{
										Layer: game.LayerType,
										AddTypes: []types.Card{
											types.Creature,
										},
										AddSubtypes: []types.Sub{
											types.Robot,
										},
									},
									{
										Layer: game.LayerPowerToughnessSet,
										SetPower: opt.Val(game.PT{
											Value: 0,
										}),
										SetToughness: opt.Val(game.PT{
											Value: 0,
										}),
									},
								},
								Duration: game.DurationPermanent,
							},
							Condition: opt.Val(game.EffectCondition{
								Text:          "it isn't a creature",
								TargetIndex:   0,
								PermanentType: opt.Val(types.Creature),
								Negate:        true,
							}),
						},
					},
				}.Ability(),
			},
		},
		TriggeredAbilities: []game.TriggeredAbilityBody{
			{
				Text: `
					Whenever an artifact you control is put into a graveyard from the battlefield, if it had counters on it, put those counters on up to one target artifact or creature you control.
				`,
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhenever,
					Pattern: game.TriggerPattern{
						Event:      game.EventZoneChanged,
						Controller: game.TriggerControllerYou,
						RequirePermanentTypes: []types.Card{
							types.Artifact,
						},
						MatchFromZone: true,
						FromZone:      zone.Battlefield,
						MatchToZone:   true,
						ToZone:        zone.Graveyard,
					},
					InterveningIf:                          "it had counters on it",
					InterveningIfEventPermanentHadCounters: true,
				},
				Content: game.Mode{
					Targets: []game.TargetSpec{
						{
							MinTargets: 0,
							MaxTargets: 1,
							Constraint: "artifact or creature you control",
						},
					},
					Sequence: []game.Instruction{
						{
							Primitive: game.MoveCounters{
								TargetIndex: 0,
								Source: game.CounterSourceSpec{
									Kind: game.CounterSourceEventPermanent,
								},
							},
							Description: "move all counters from the triggering artifact to target",
						},
					},
				}.Ability(),
			},
		},
	},
}
