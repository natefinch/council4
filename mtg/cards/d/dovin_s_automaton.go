package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DovinSAutomaton is the card definition for Dovin's Automaton.
//
// Type: Artifact Creature — Homunculus
// Cost: {4}
//
// Oracle text:
//
//	As long as you control a Dovin planeswalker, this creature gets +2/+2 and has vigilance. (Attacking doesn't cause it to tap.)
var DovinSAutomaton = newDovinSAutomaton

func newDovinSAutomaton() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Dovin's Automaton",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Homunculus},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "As long as you control a Dovin planeswalker, this creature gets +2/+2 and has vigilance. (Attacking doesn't cause it to tap.)",
					Condition: opt.Val(game.Condition{
						Text: "As long as you control a Dovin planeswalker",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}, SubtypesAny: []types.Sub{types.Dovin}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Vigilance},
						},
					},
				},
			},
			OracleText: `
			As long as you control a Dovin planeswalker, this creature gets +2/+2 and has vigilance. (Attacking doesn't cause it to tap.)
		`,
		},
	}
}
