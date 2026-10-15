package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VintaraSnapper is the card definition for Vintara Snapper.
//
// Type: Creature — Turtle
// Cost: {G}{G}
//
// Oracle text:
//
//	This creature has shroud as long as you control no untapped lands. (It can't be the target of spells or abilities.)
var VintaraSnapper = newVintaraSnapper

func newVintaraSnapper() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Vintara Snapper",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Turtle},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature has shroud as long as you control no untapped lands. (It can't be the target of spells or abilities.)",
					Condition: opt.Val(game.Condition{
						Text:   "as long as you control no untapped lands",
						Negate: true,
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Land}, Tapped: game.TriFalse},
							MinCount:  1,
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Shroud},
						},
					},
				},
			},
			OracleText: `
			This creature has shroud as long as you control no untapped lands. (It can't be the target of spells or abilities.)
		`,
		},
	}
}
