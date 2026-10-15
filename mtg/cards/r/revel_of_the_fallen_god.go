package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RevelOfTheFallenGod is the card definition for Revel of the Fallen God.
//
// Type: Sorcery
// Cost: {3}{R}{R}{G}{G}
//
// Oracle text:
//
//	Create four 2/2 red and green Satyr creature tokens with haste.
var RevelOfTheFallenGod = newRevelOfTheFallenGod

func newRevelOfTheFallenGod() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Revel of the Fallen God",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green, color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(4),
							Source: game.TokenDef(revelOfTheFallenGodToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create four 2/2 red and green Satyr creature tokens with haste.
		`,
		},
	}
}

var revelOfTheFallenGodToken = newRevelOfTheFallenGodToken()

func newRevelOfTheFallenGodToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Satyr",
			Colors:    []color.Color{color.Red, color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Satyr},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
		},
	}
}
