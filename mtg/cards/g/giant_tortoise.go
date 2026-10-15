package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GiantTortoise is the card definition for Giant Tortoise.
//
// Type: Creature — Turtle
// Cost: {1}{U}
//
// Oracle text:
//
//	This creature gets +0/+3 as long as it's untapped.
var GiantTortoise = newGiantTortoise

func newGiantTortoise() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Giant Tortoise",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Turtle},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature gets +0/+3 as long as it's untapped.",
					Condition: opt.Val(game.Condition{
						Text:              "as long as it's untapped",
						Object:            opt.Val(game.SourcePermanentReference()),
						ObjectMatches:     opt.Val(game.Selection{Tapped: game.TriFalse}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							ToughnessDelta: 3,
						},
					},
				},
			},
			OracleText: `
			This creature gets +0/+3 as long as it's untapped.
		`,
		},
	}
}
