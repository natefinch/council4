package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ImperialOath is the card definition for Imperial Oath.
//
// Type: Sorcery
// Cost: {5}{W}
//
// Oracle text:
//
//	Create three 2/2 white Samurai creature tokens with vigilance. Scry 3.
var ImperialOath = newImperialOath

func newImperialOath() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Imperial Oath",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(3),
							Source: game.TokenDef(imperialOathToken),
						},
					},
					{
						Primitive: game.Scry{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create three 2/2 white Samurai creature tokens with vigilance. Scry 3.
		`,
		},
	}
}

var imperialOathToken = newImperialOathToken()

func newImperialOathToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Samurai",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Samurai},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
		},
	}
}
