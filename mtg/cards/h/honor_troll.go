package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HonorTroll is the card definition for Honor Troll.
//
// Type: Creature — Troll Druid
// Cost: {2}{G}
//
// Oracle text:
//
//	Vigilance
//	If you would gain life, you gain that much life plus 1 instead.
//	This creature gets +2/+1 as long as you have 25 or more life.
var HonorTroll = newHonorTroll

func newHonorTroll() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Honor Troll",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Troll, types.Druid},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.StaticAbility{
					Text: "This creature gets +2/+1 as long as you have 25 or more life.",
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
							ToughnessDelta: 1,
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.LifeGainReplacement("If you would gain life, you gain that much life plus 1 instead.", 1, 1),
			},
			OracleText: `
			Vigilance
			If you would gain life, you gain that much life plus 1 instead.
			This creature gets +2/+1 as long as you have 25 or more life.
		`,
		},
	}
}
