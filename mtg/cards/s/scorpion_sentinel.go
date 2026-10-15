package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ScorpionSentinel is the card definition for Scorpion Sentinel.
//
// Type: Artifact Creature — Robot Scorpion
// Cost: {1}{U}
//
// Oracle text:
//
//	As long as you control seven or more lands, this creature gets +3/+0.
var ScorpionSentinel = newScorpionSentinel

func newScorpionSentinel() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Scorpion Sentinel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Robot, types.Scorpion},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "As long as you control seven or more lands, this creature gets +3/+0.",
					Condition: opt.Val(game.Condition{
						Text: "As long as you control seven or more lands",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
							MinCount:  7,
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     3,
						},
					},
				},
			},
			OracleText: `
			As long as you control seven or more lands, this creature gets +3/+0.
		`,
		},
	}
}
