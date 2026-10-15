package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AuriokEdgewright is the card definition for Auriok Edgewright.
//
// Type: Creature — Human Soldier
// Cost: {W}{W}
//
// Oracle text:
//
//	Metalcraft — This creature has double strike as long as you control three or more artifacts.
var AuriokEdgewright = newAuriokEdgewright

func newAuriokEdgewright() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Auriok Edgewright",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Metalcraft — This creature has double strike as long as you control three or more artifacts.",
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
							AddKeywords:    []game.Keyword{game.DoubleStrike},
						},
					},
				},
			},
			OracleText: `
			Metalcraft — This creature has double strike as long as you control three or more artifacts.
		`,
		},
	}
}
