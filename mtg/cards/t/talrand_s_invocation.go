package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TalrandSInvocation is the card definition for Talrand's Invocation.
//
// Type: Sorcery
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	Create two 2/2 blue Drake creature tokens with flying.
var TalrandSInvocation = newTalrandSInvocation

func newTalrandSInvocation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Talrand's Invocation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(2),
							Source: game.TokenDef(talrandSInvocationToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create two 2/2 blue Drake creature tokens with flying.
		`,
		},
	}
}

var talrandSInvocationToken = newTalrandSInvocationToken()

func newTalrandSInvocationToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Drake",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Drake},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
