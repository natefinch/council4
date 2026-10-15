package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TakeUpArms is the card definition for Take Up Arms.
//
// Type: Instant
// Cost: {4}{W}
//
// Oracle text:
//
//	Create three 1/1 white Warrior creature tokens.
var TakeUpArms = newTakeUpArms

func newTakeUpArms() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Take Up Arms",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(3),
							Source: game.TokenDef(takeUpArmsToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create three 1/1 white Warrior creature tokens.
		`,
		},
	}
}

var takeUpArmsToken = newTakeUpArmsToken()

func newTakeUpArmsToken() *game.CardDef {
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
