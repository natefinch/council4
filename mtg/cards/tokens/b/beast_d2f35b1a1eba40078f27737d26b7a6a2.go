package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BeastTokend2f35b1a1eba40078f27737d26b7a6a2 is the card definition for Beast.
//
// Type: Token Creature — Beast
//
// Oracle text:
//   This creature can't attack or block alone.

// BeastTokend2f35b1a1eba40078f27737d26b7a6a2 is the card definition for Beast.
var BeastTokend2f35b1a1eba40078f27737d26b7a6a2 = newBeastTokend2f35b1a1eba40078f27737d26b7a6a2()

func newBeastTokend2f35b1a1eba40078f27737d26b7a6a2() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:      "Beast",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.CantAttackOrBlockAloneStaticBody,
			},
			OracleText: `
			This creature can't attack or block alone.
		`,
		},
	}
}
