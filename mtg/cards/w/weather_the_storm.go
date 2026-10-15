package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WeatherTheStorm is the card definition for Weather the Storm.
//
// Type: Instant
// Cost: {1}{G}
//
// Oracle text:
//
//	You gain 3 life.
//	Storm (When you cast this spell, copy it for each spell cast before it this turn.)
var WeatherTheStorm = newWeatherTheStorm

func newWeatherTheStorm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Weather the Storm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StormStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			You gain 3 life.
			Storm (When you cast this spell, copy it for each spell cast before it this turn.)
		`,
		},
	}
}
