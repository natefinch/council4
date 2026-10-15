package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DeathmaskNezumi is the card definition for Deathmask Nezumi.
//
// Type: Creature — Rat Shaman
// Cost: {2}{B}
//
// Oracle text:
//
//	As long as you have seven or more cards in hand, this creature gets +2/+1 and has fear. (It can't be blocked except by artifact creatures and/or black creatures.)
var DeathmaskNezumi = newDeathmaskNezumi

func newDeathmaskNezumi() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Deathmask Nezumi",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Rat, types.Shaman},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "As long as you have seven or more cards in hand, this creature gets +2/+1 and has fear. (It can't be blocked except by artifact creatures and/or black creatures.)",
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
							AddKeywords:    []game.Keyword{game.Fear},
						},
					},
				},
			},
			OracleText: `
			As long as you have seven or more cards in hand, this creature gets +2/+1 and has fear. (It can't be blocked except by artifact creatures and/or black creatures.)
		`,
		},
	}
}
