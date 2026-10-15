package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThornwatchScarecrow is the card definition for Thornwatch Scarecrow.
//
// Type: Artifact Creature — Scarecrow
// Cost: {6}
//
// Oracle text:
//
//	This creature has wither as long as you control a green creature. (It deals damage to creatures in the form of -1/-1 counters.)
//	This creature has vigilance as long as you control a white creature.
var ThornwatchScarecrow = newThornwatchScarecrow

func newThornwatchScarecrow() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Thornwatch Scarecrow",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Scarecrow},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature has wither as long as you control a green creature. (It deals damage to creatures in the form of -1/-1 counters.)",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control a green creature",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.Green}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Wither},
						},
					},
				},
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
			},
			OracleText: `
			This creature has wither as long as you control a green creature. (It deals damage to creatures in the form of -1/-1 counters.)
			This creature has vigilance as long as you control a white creature.
		`,
		},
	}
}
