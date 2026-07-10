package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// KnightedMyr is the card definition for Knighted Myr.
//
// Type: Artifact Creature — Myr Knight
// Cost: {2}{W}
//
// Oracle text:
//
//	{2}{W}: Adapt 1. (If this creature has no +1/+1 counters on it, put a +1/+1 counter on it.)
//	Whenever one or more +1/+1 counters are put on this creature, it gains double strike until end of turn.
var KnightedMyr = newKnightedMyr

func newKnightedMyr() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Knighted Myr",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Myr, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{2}{W}: Adapt 1. (If this creature has no +1/+1 counters on it, put a +1/+1 counter on it.)",
					ManaCost:       opt.Val(cost.Mana{cost.O(2), cost.W}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Adapt{
									Object: game.SourcePermanentReference(),
									Amount: game.Fixed(1),
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventCountersAdded,
							Source:           game.TriggerSourceSelf,
							OneOrMore:        true,
							MatchCounterKind: true,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.SourceCardPermanentReference()),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.DoubleStrike,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{2}{W}: Adapt 1. (If this creature has no +1/+1 counters on it, put a +1/+1 counter on it.)
			Whenever one or more +1/+1 counters are put on this creature, it gains double strike until end of turn.
		`,
		},
	}
}
