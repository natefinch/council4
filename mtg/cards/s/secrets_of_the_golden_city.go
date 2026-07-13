package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SecretsOfTheGoldenCity is the card definition for Secrets of the Golden City.
//
// Type: Sorcery
// Cost: {1}{U}{U}
//
// Oracle text:
//
//	Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)
//	Draw two cards. If you have the city's blessing, draw three cards instead.
var SecretsOfTheGoldenCity = newSecretsOfTheGoldenCity

func newSecretsOfTheGoldenCity() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Secrets of the Golden City",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.GainCityBlessing{},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:                    true,
								ControllerHasCityBlessing: true,
							}),
						}),
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								ControllerHasCityBlessing: true,
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)
			Draw two cards. If you have the city's blessing, draw three cards instead.
		`,
		},
	}
}
