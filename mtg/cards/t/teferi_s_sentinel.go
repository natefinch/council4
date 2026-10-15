package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TeferiSSentinel is the card definition for Teferi's Sentinel.
//
// Type: Artifact Creature — Golem
// Cost: {5}
//
// Oracle text:
//
//	As long as you control a Teferi planeswalker, this creature gets +4/+0.
var TeferiSSentinel = newTeferiSSentinel

func newTeferiSSentinel() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Teferi's Sentinel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "As long as you control a Teferi planeswalker, this creature gets +4/+0.",
					Condition: opt.Val(game.Condition{
						Text: "As long as you control a Teferi planeswalker",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}, SubtypesAny: []types.Sub{types.Teferi}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     4,
						},
					},
				},
			},
			OracleText: `
			As long as you control a Teferi planeswalker, this creature gets +4/+0.
		`,
		},
	}
}
