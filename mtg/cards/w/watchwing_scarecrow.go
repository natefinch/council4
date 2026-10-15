package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WatchwingScarecrow is the card definition for Watchwing Scarecrow.
//
// Type: Artifact Creature — Scarecrow
// Cost: {4}
//
// Oracle text:
//
//	This creature has vigilance as long as you control a white creature.
//	This creature has flying as long as you control a blue creature.
var WatchwingScarecrow = newWatchwingScarecrow

func newWatchwingScarecrow() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Watchwing Scarecrow",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Scarecrow},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature has vigilance as long as you control a white creature.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control a white creature",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.White}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Vigilance},
						},
					},
				},
				game.StaticAbility{
					Text: "This creature has flying as long as you control a blue creature.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control a blue creature",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.Blue}},
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
			This creature has vigilance as long as you control a white creature.
			This creature has flying as long as you control a blue creature.
		`,
		},
	}
}
