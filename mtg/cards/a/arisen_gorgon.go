package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArisenGorgon is the card definition for Arisen Gorgon.
//
// Type: Creature — Zombie Gorgon
// Cost: {1}{B}{B}
//
// Oracle text:
//
//	This creature has deathtouch as long as you control a Liliana planeswalker. (Any amount of damage this deals to a creature is enough to destroy it.)
var ArisenGorgon = newArisenGorgon

func newArisenGorgon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Arisen Gorgon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Gorgon},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature has deathtouch as long as you control a Liliana planeswalker. (Any amount of damage this deals to a creature is enough to destroy it.)",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control a Liliana planeswalker",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}, SubtypesAny: []types.Sub{types.Liliana}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Deathtouch},
						},
					},
				},
			},
			OracleText: `
			This creature has deathtouch as long as you control a Liliana planeswalker. (Any amount of damage this deals to a creature is enough to destroy it.)
		`,
		},
	}
}
