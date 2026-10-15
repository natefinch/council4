package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CarrionCall is the card definition for Carrion Call.
//
// Type: Instant
// Cost: {3}{G}
//
// Oracle text:
//
//	Create two 1/1 green Phyrexian Insect creature tokens with infect. (They deal damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)
var CarrionCall = newCarrionCall

func newCarrionCall() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Carrion Call",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(2),
							Source: game.TokenDef(carrionCallToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create two 1/1 green Phyrexian Insect creature tokens with infect. (They deal damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)
		`,
		},
	}
}

var carrionCallToken = newCarrionCallToken()

func newCarrionCallToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Phyrexian Insect",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Insect},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.InfectStaticBody,
			},
		},
	}
}
