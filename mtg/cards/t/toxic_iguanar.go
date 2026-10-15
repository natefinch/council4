package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ToxicIguanar is the card definition for Toxic Iguanar.
//
// Type: Creature — Lizard
// Cost: {R}
//
// Oracle text:
//
//	This creature has deathtouch as long as you control a green permanent. (Any amount of damage a creature with deathtouch deals to a creature is enough to destroy it.)
var ToxicIguanar = newToxicIguanar

func newToxicIguanar() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Toxic Iguanar",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Lizard},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature has deathtouch as long as you control a green permanent. (Any amount of damage a creature with deathtouch deals to a creature is enough to destroy it.)",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control a green permanent",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{ColorsAny: []color.Color{color.Green}},
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
			This creature has deathtouch as long as you control a green permanent. (Any amount of damage a creature with deathtouch deals to a creature is enough to destroy it.)
		`,
		},
	}
}
