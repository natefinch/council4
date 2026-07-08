package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JaceSPhantasm is the card definition for Jace's Phantasm.
//
// Type: Creature — Illusion
// Cost: {U}
//
// Oracle text:
//
//	Flying
//	This creature gets +4/+4 as long as an opponent has ten or more cards in their graveyard.
var JaceSPhantasm = newJaceSPhantasm

func newJaceSPhantasm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Jace's Phantasm",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Illusion},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateAnyOpponentGraveyardCardCount, Op: compare.GreaterOrEqual, Value: 10}},
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							AffectedSource: true,
							PowerDelta:     4,
							ToughnessDelta: 4,
						},
					},
				},
			},
			OracleText: `
			Flying
			This creature gets +4/+4 as long as an opponent has ten or more cards in their graveyard.
		`,
		},
	}
}
