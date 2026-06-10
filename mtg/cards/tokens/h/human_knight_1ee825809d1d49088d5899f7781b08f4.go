package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Human Knight
//
// Type: Token Creature — Human Knight
//
// Oracle text:
//   Trample, haste

// HumanKnightToken1ee825809d1d49088d5899f7781b08f4 is the card definition for Human Knight.
var HumanKnightToken1ee825809d1d49088d5899f7781b08f4 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Human Knight",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Human, types.Knight},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.TrampleStaticBody,
			game.HasteStaticBody,
		},
		OracleText: `
			Trample, haste
		`,
	},
}
