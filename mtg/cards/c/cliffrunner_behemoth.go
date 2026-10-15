package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CliffrunnerBehemoth is the card definition for Cliffrunner Behemoth.
//
// Type: Creature — Rhino Beast
// Cost: {3}{G}
//
// Oracle text:
//
//	This creature has haste as long as you control a red permanent.
//	This creature has lifelink as long as you control a white permanent.
var CliffrunnerBehemoth = newCliffrunnerBehemoth

func newCliffrunnerBehemoth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Cliffrunner Behemoth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Rhino, types.Beast},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature has haste as long as you control a red permanent.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control a red permanent",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{ColorsAny: []color.Color{color.Red}},
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
					Text: "This creature has lifelink as long as you control a white permanent.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control a white permanent",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{ColorsAny: []color.Color{color.White}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Lifelink},
						},
					},
				},
			},
			OracleText: `
			This creature has haste as long as you control a red permanent.
			This creature has lifelink as long as you control a white permanent.
		`,
		},
	}
}
