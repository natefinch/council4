package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IslebackSpawn is the card definition for Isleback Spawn.
//
// Type: Creature — Kraken
// Cost: {5}{U}{U}
//
// Oracle text:
//
//	Shroud (This creature can't be the target of spells or abilities.)
//	This creature gets +4/+8 as long as a library has twenty or fewer cards in it.
var IslebackSpawn = newIslebackSpawn

func newIslebackSpawn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Isleback Spawn",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kraken},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 8}),
			StaticAbilities: []game.StaticAbility{
				game.ShroudStaticBody,
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateMinPlayerLibrarySize, Op: compare.LessOrEqual, Value: 20}},
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							AffectedSource: true,
							PowerDelta:     4,
							ToughnessDelta: 8,
						},
					},
				},
			},
			OracleText: `
			Shroud (This creature can't be the target of spells or abilities.)
			This creature gets +4/+8 as long as a library has twenty or fewer cards in it.
		`,
		},
	}
}
