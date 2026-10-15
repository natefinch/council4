package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThrabenFoulbloods is the card definition for Thraben Foulbloods.
//
// Type: Creature — Zombie Dog
// Cost: {2}{B}
//
// Oracle text:
//
//	Delirium — This creature gets +1/+1 and has menace as long as there are four or more card types among cards in your graveyard. (A creature with menace can't be blocked except by two or more creatures.)
var ThrabenFoulbloods = newThrabenFoulbloods

func newThrabenFoulbloods() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Thraben Foulbloods",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Dog},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Delirium — This creature gets +1/+1 and has menace as long as there are four or more card types among cards in your graveyard. (A creature with menace can't be blocked except by two or more creatures.)",
					Condition: opt.Val(game.Condition{
						Text: "as long as there are four or more card types among cards in your graveyard",
						Aggregates: []game.AggregateComparison{
							game.AggregateComparison{
								Aggregate: game.AggregateControllerGraveyardCardTypeCount,
								Op:        compare.GreaterOrEqual,
								Value:     4,
							},
						},
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Menace},
						},
					},
				},
			},
			OracleText: `
			Delirium — This creature gets +1/+1 and has menace as long as there are four or more card types among cards in your graveyard. (A creature with menace can't be blocked except by two or more creatures.)
		`,
		},
	}
}
