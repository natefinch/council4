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

// AnuridBarkripper is the card definition for Anurid Barkripper.
//
// Type: Creature — Frog Beast
// Cost: {1}{G}{G}
//
// Oracle text:
//
//	Threshold — This creature gets +2/+2 as long as there are seven or more cards in your graveyard.
var AnuridBarkripper = newAnuridBarkripper

func newAnuridBarkripper() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Anurid Barkripper",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Frog, types.Beast},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Threshold — This creature gets +2/+2 as long as there are seven or more cards in your graveyard.",
					Condition: opt.Val(game.Condition{
						Text: "as long as there are seven or more cards in your graveyard",
						Aggregates: []game.AggregateComparison{
							game.AggregateComparison{
								Aggregate: game.AggregateControllerGraveyardCardCount,
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
							ToughnessDelta: 2,
						},
					},
				},
			},
			OracleText: `
			Threshold — This creature gets +2/+2 as long as there are seven or more cards in your graveyard.
		`,
		},
	}
}
