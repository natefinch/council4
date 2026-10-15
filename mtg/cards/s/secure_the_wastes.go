package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SecureTheWastes is the card definition for Secure the Wastes.
//
// Type: Instant
// Cost: {X}{W}
//
// Oracle text:
//
//	Create X 1/1 white Warrior creature tokens.
var SecureTheWastes = newSecureTheWastes

func newSecureTheWastes() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Secure the Wastes",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Source: game.TokenDef(secureTheWastesToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create X 1/1 white Warrior creature tokens.
		`,
		},
	}
}

var secureTheWastesToken = newSecureTheWastesToken()

func newSecureTheWastesToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Warrior",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Warrior},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
