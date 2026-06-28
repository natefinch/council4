package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NimanaSkitterSneak is the card definition for Nimana Skitter-Sneak.
//
// Type: Creature — Human Rogue
// Cost: {3}{B}
//
// Oracle text:
//
//	As long as an opponent has eight or more cards in their graveyard, this creature gets +1/+0 and has menace. (It can't be blocked except by two or more creatures.)
var NimanaSkitterSneak = newNimanaSkitterSneak()

func newNimanaSkitterSneak() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Nimana Skitter-Sneak",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Rogue},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateAnyOpponentGraveyardCardCount, Op: compare.GreaterOrEqual, Value: 8}},
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							AffectedSource: true,
							PowerDelta:     1,
						},
						game.ContinuousEffect{
							Layer:          game.LayerAbility,
							AffectedSource: true,
							AddKeywords: []game.Keyword{
								game.Menace,
							},
						},
					},
				},
			},
			OracleText: `
			As long as an opponent has eight or more cards in their graveyard, this creature gets +1/+0 and has menace. (It can't be blocked except by two or more creatures.)
		`,
		},
	}
}
