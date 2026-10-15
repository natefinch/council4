package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DivinityOfPride is the card definition for Divinity of Pride.
//
// Type: Creature — Spirit Avatar
// Cost: {W/B}{W/B}{W/B}{W/B}{W/B}
//
// Oracle text:
//
//	Flying, lifelink
//	This creature gets +4/+4 as long as you have 25 or more life.
var DivinityOfPride = newDivinityOfPride

func newDivinityOfPride() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Divinity of Pride",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.W, mana.B),
				cost.HybridMana(mana.W, mana.B),
				cost.HybridMana(mana.W, mana.B),
				cost.HybridMana(mana.W, mana.B),
				cost.HybridMana(mana.W, mana.B),
			}),
			Colors:    []color.Color{color.Black, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit, types.Avatar},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.LifelinkStaticBody,
				game.StaticAbility{
					Text: "This creature gets +4/+4 as long as you have 25 or more life.",
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
							PowerDelta:     4,
							ToughnessDelta: 4,
						},
					},
				},
			},
			OracleText: `
			Flying, lifelink
			This creature gets +4/+4 as long as you have 25 or more life.
		`,
		},
	}
}
