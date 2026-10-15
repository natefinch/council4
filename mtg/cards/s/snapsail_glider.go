package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SnapsailGlider is the card definition for Snapsail Glider.
//
// Type: Artifact Creature — Construct
// Cost: {3}
//
// Oracle text:
//
//	Metalcraft — This creature has flying as long as you control three or more artifacts.
var SnapsailGlider = newSnapsailGlider

func newSnapsailGlider() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Snapsail Glider",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Metalcraft — This creature has flying as long as you control three or more artifacts.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control three or more artifacts",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
							MinCount:  3,
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
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
			Metalcraft — This creature has flying as long as you control three or more artifacts.
		`,
		},
	}
}
