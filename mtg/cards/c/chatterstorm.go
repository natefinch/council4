package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Chatterstorm is the card definition for Chatterstorm.
//
// Type: Sorcery
// Cost: {1}{G}
//
// Oracle text:
//
//	Create a 1/1 green Squirrel creature token.
//	Storm (When you cast this spell, copy it for each spell cast before it this turn.)
var Chatterstorm = newChatterstorm

func newChatterstorm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Chatterstorm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StormStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(chatterstormToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 1/1 green Squirrel creature token.
			Storm (When you cast this spell, copy it for each spell cast before it this turn.)
		`,
		},
	}
}

var chatterstormToken = newChatterstormToken()

func newChatterstormToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Squirrel",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Squirrel},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
