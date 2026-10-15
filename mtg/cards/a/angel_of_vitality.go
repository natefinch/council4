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

// AngelOfVitality is the card definition for Angel of Vitality.
//
// Type: Creature — Angel
// Cost: {2}{W}
//
// Oracle text:
//
//	Flying
//	If you would gain life, you gain that much life plus 1 instead.
//	This creature gets +2/+2 as long as you have 25 or more life.
var AngelOfVitality = newAngelOfVitality

func newAngelOfVitality() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Angel of Vitality",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					Text: "This creature gets +2/+2 as long as you have 25 or more life.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you have 25 or more life",
						Aggregates: []game.AggregateComparison{
							game.AggregateComparison{
								Aggregate: game.AggregateControllerLife,
								Op:        compare.GreaterOrEqual,
								Value:     25,
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.LifeGainReplacement("If you would gain life, you gain that much life plus 1 instead.", 1, 1),
			},
			OracleText: `
			Flying
			If you would gain life, you gain that much life plus 1 instead.
			This creature gets +2/+2 as long as you have 25 or more life.
		`,
		},
	}
}
