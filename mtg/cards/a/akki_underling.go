package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AkkiUnderling is the card definition for Akki Underling.
//
// Type: Creature — Goblin Warrior
// Cost: {1}{R}
//
// Oracle text:
//
//	As long as you have seven or more cards in hand, this creature gets +2/+1 and has first strike.
var AkkiUnderling = newAkkiUnderling

func newAkkiUnderling() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Akki Underling",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "As long as you have seven or more cards in hand, this creature gets +2/+1 and has first strike.",
					Condition: opt.Val(game.Condition{
						Text: "As long as you have seven or more cards in hand",
						Aggregates: []game.AggregateComparison{
							game.AggregateComparison{
								Aggregate: game.AggregateControllerHandSize,
								Op:        compare.GreaterOrEqual,
								Value:     7,
							},
						},
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.FirstStrike},
						},
					},
				},
			},
			OracleText: `
			As long as you have seven or more cards in hand, this creature gets +2/+1 and has first strike.
		`,
		},
	}
}
