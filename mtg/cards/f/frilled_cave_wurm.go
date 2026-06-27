package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FrilledCaveWurm is the card definition for Frilled Cave-Wurm.
//
// Type: Creature — Salamander Wurm
// Cost: {3}{U}
//
// Oracle text:
//
//	Descend 4 — This creature gets +2/+0 as long as there are four or more permanent cards in your graveyard.
var FrilledCaveWurm = newFrilledCaveWurm()

func newFrilledCaveWurm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Frilled Cave-Wurm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Salamander, types.Wurm},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerGraveyardPermanentCardCount, Op: compare.GreaterOrEqual, Value: 4}},
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							AffectedSource: true,
							PowerDelta:     2,
						},
					},
				},
			},
			OracleText: `
			Descend 4 — This creature gets +2/+0 as long as there are four or more permanent cards in your graveyard.
		`,
		},
	}
}
