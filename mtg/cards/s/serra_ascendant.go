package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SerraAscendant is the card definition for Serra Ascendant.
//
// Type: Creature — Human Monk
// Cost: {W}
//
// Oracle text:
//
//	Lifelink (Damage dealt by this creature also causes you to gain that much life.)
//	As long as you have 30 or more life, this creature gets +5/+5 and has flying.
var SerraAscendant = newSerraAscendant

func newSerraAscendant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Serra Ascendant",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Monk},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.LifelinkStaticBody,
				game.StaticAbility{
					Text: "As long as you have 30 or more life, this creature gets +5/+5 and has flying.",
					Condition: opt.Val(game.Condition{
						Text: "As long as you have 30 or more life",
						Aggregates: []game.AggregateComparison{
							game.AggregateComparison{
								Aggregate: game.AggregateControllerLife,
								Op:        compare.GreaterOrEqual,
								Value:     30,
							},
						},
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     5,
							ToughnessDelta: 5,
						},
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Flying},
						},
					},
				},
			},
			OracleText: `
			Lifelink (Damage dealt by this creature also causes you to gain that much life.)
			As long as you have 30 or more life, this creature gets +5/+5 and has flying.
		`,
		},
	}
}
