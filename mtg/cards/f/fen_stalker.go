package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FenStalker is the card definition for Fen Stalker.
//
// Type: Creature — Nightstalker
// Cost: {3}{B}
//
// Oracle text:
//
//	This creature has fear as long as you control no untapped lands. (It can't be blocked except by artifact creatures and/or black creatures.)
var FenStalker = newFenStalker

func newFenStalker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Fen Stalker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Nightstalker},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature has fear as long as you control no untapped lands. (It can't be blocked except by artifact creatures and/or black creatures.)",
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
							AddKeywords:    []game.Keyword{game.Fear},
						},
					},
				},
			},
			OracleText: `
			This creature has fear as long as you control no untapped lands. (It can't be blocked except by artifact creatures and/or black creatures.)
		`,
		},
	}
}
