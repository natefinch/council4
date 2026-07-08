package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PrimordialHydra is the card definition for Primordial Hydra.
//
// Type: Creature — Hydra
// Cost: {X}{G}{G}
//
// Oracle text:
//
//	This creature enters with X +1/+1 counters on it.
//	At the beginning of your upkeep, double the number of +1/+1 counters on this creature.
//	This creature has trample as long as it has ten or more +1/+1 counters on it.
var PrimordialHydra = newPrimordialHydra

func newPrimordialHydra() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Primordial Hydra",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Hydra},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Object:        opt.Val(game.SourcePermanentReference()),
						ObjectMatches: opt.Val(game.Selection{RequiredCounterCount: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 10}), RequiredCounter: counter.PlusOnePlusOne}),
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerAbility,
							AffectedSource: true,
							AddKeywords: []game.Keyword{
								game.Trample,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
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
								Primitive: game.AddCounter{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:        game.DynamicAmountObjectCounters,
										CounterKind: counter.PlusOnePlusOne,
										Object:      game.SourcePermanentReference(),
									}),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with X +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, AmountFromX: true}),
			},
			OracleText: `
			This creature enters with X +1/+1 counters on it.
			At the beginning of your upkeep, double the number of +1/+1 counters on this creature.
			This creature has trample as long as it has ten or more +1/+1 counters on it.
		`,
		},
	}
}
