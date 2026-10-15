package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SummitApes is the card definition for Summit Apes.
//
// Type: Creature — Ape
// Cost: {3}{G}
//
// Oracle text:
//
//	As long as you control a Mountain, this creature has menace. (It can't be blocked except by two or more creatures.)
var SummitApes = newSummitApes

func newSummitApes() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Summit Apes",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ape},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "As long as you control a Mountain, this creature has menace. (It can't be blocked except by two or more creatures.)",
					Condition: opt.Val(game.Condition{
						Text: "As long as you control a Mountain",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{SubtypesAny: []types.Sub{types.Mountain}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Menace},
						},
					},
				},
			},
			OracleText: `
			As long as you control a Mountain, this creature has menace. (It can't be blocked except by two or more creatures.)
		`,
		},
	}
}
