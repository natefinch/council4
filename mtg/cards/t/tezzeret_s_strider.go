package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TezzeretSStrider is the card definition for Tezzeret's Strider.
//
// Type: Artifact Creature — Golem
// Cost: {3}
//
// Oracle text:
//
//	As long as you control a Tezzeret planeswalker, this creature has menace. (It can't be blocked except by two or more creatures.)
var TezzeretSStrider = newTezzeretSStrider

func newTezzeretSStrider() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Tezzeret's Strider",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "As long as you control a Tezzeret planeswalker, this creature has menace. (It can't be blocked except by two or more creatures.)",
					Condition: opt.Val(game.Condition{
						Text: "As long as you control a Tezzeret planeswalker",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}, SubtypesAny: []types.Sub{types.Tezzeret}},
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
			As long as you control a Tezzeret planeswalker, this creature has menace. (It can't be blocked except by two or more creatures.)
		`,
		},
	}
}
