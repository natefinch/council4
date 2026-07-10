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

// DreamdrinkerVampire is the card definition for Dreamdrinker Vampire.
//
// Type: Creature — Vampire
// Cost: {1}{B}
//
// Oracle text:
//
//	Lifelink
//	{1}{B}: Adapt 1. (If this creature has no +1/+1 counters on it, put a +1/+1 counter on it.)
//	Whenever one or more +1/+1 counters are put on this creature, it gains menace until end of turn.
var DreamdrinkerVampire = newDreamdrinkerVampire

func newDreamdrinkerVampire() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Dreamdrinker Vampire",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.LifelinkStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{1}{B}: Adapt 1. (If this creature has no +1/+1 counters on it, put a +1/+1 counter on it.)",
					ManaCost:       opt.Val(cost.Mana{cost.O(1), cost.B}),
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
												game.Menace,
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
			Lifelink
			{1}{B}: Adapt 1. (If this creature has no +1/+1 counters on it, put a +1/+1 counter on it.)
			Whenever one or more +1/+1 counters are put on this creature, it gains menace until end of turn.
		`,
		},
	}
}
