package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HeadstrongBrute is the card definition for Headstrong Brute.
//
// Type: Creature — Orc Pirate
// Cost: {2}{R}
//
// Oracle text:
//
//	This creature can't block.
//	This creature has menace as long as you control another Pirate.
var HeadstrongBrute = newHeadstrongBrute

func newHeadstrongBrute() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Headstrong Brute",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Orc, types.Pirate},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.CantBlockStaticBody,
				game.StaticAbility{
					Text: "This creature has menace as long as you control another Pirate.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control another Pirate",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{SubtypesAny: []types.Sub{types.Pirate}, ExcludeSource: true},
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
			This creature can't block.
			This creature has menace as long as you control another Pirate.
		`,
		},
	}
}
