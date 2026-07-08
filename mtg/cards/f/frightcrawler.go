package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Frightcrawler is the card definition for Frightcrawler.
//
// Type: Creature — Horror
// Cost: {1}{B}
//
// Oracle text:
//
//	Fear (This creature can't be blocked except by artifact creatures and/or black creatures.)
//	Threshold — As long as there are seven or more cards in your graveyard, this creature gets +2/+2 and can't block.
var Frightcrawler = newFrightcrawler

func newFrightcrawler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Frightcrawler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Horror},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FearStaticBody,
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerGraveyardCardCount, Op: compare.GreaterOrEqual, Value: 7}},
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							AffectedSource: true,
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
					},
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBlock,
							AffectedSource: true,
						},
					},
				},
			},
			OracleText: `
			Fear (This creature can't be blocked except by artifact creatures and/or black creatures.)
			Threshold — As long as there are seven or more cards in your graveyard, this creature gets +2/+2 and can't block.
		`,
		},
	}
}
