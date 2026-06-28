package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EchoOfDusk is the card definition for Echo of Dusk.
//
// Type: Creature — Vampire Spirit
// Cost: {1}{B}
//
// Oracle text:
//
//	Descend 4 — As long as there are four or more permanent cards in your graveyard, this creature gets +1/+1 and has lifelink.
var EchoOfDusk = newEchoOfDusk()

func newEchoOfDusk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Echo of Dusk",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Spirit},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerGraveyardPermanentCardCount, Op: compare.GreaterOrEqual, Value: 4}},
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							AffectedSource: true,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Layer:          game.LayerAbility,
							AffectedSource: true,
							AddKeywords: []game.Keyword{
								game.Lifelink,
							},
						},
					},
				},
			},
			OracleText: `
			Descend 4 — As long as there are four or more permanent cards in your graveyard, this creature gets +1/+1 and has lifelink.
		`,
		},
	}
}
