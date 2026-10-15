package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BlazethornScarecrow is the card definition for Blazethorn Scarecrow.
//
// Type: Artifact Creature — Scarecrow
// Cost: {5}
//
// Oracle text:
//
//	This creature has haste as long as you control a red creature.
//	This creature has wither as long as you control a green creature. (It deals damage to creatures in the form of -1/-1 counters.)
var BlazethornScarecrow = newBlazethornScarecrow

func newBlazethornScarecrow() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Blazethorn Scarecrow",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Scarecrow},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature has haste as long as you control a red creature.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control a red creature",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.Red}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Haste},
						},
					},
				},
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
			},
			OracleText: `
			This creature has haste as long as you control a red creature.
			This creature has wither as long as you control a green creature. (It deals damage to creatures in the form of -1/-1 counters.)
		`,
		},
	}
}
