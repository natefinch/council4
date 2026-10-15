package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CennSEnlistment is the card definition for Cenn's Enlistment.
//
// Type: Sorcery
// Cost: {3}{W}
//
// Oracle text:
//
//	Create two 1/1 white Kithkin Soldier creature tokens.
//	Retrace (You may cast this card from your graveyard by discarding a land card in addition to paying its other costs.)
var CennSEnlistment = newCennSEnlistment

func newCennSEnlistment() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Cenn's Enlistment",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.RetraceStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(2),
							Source: game.TokenDef(cennSEnlistmentToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create two 1/1 white Kithkin Soldier creature tokens.
			Retrace (You may cast this card from your graveyard by discarding a land card in addition to paying its other costs.)
		`,
		},
	}
}

var cennSEnlistmentToken = newCennSEnlistmentToken()

func newCennSEnlistmentToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Kithkin Soldier",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kithkin, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
