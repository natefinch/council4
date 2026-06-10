package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Spirit
//
// Type: Token Creature — Spirit
//
// Oracle text:
//   Flying

// SpiritToken18ef473df9584bff8045154f6adc6b04 is the card definition for Spirit.
var SpiritToken18ef473df9584bff8045154f6adc6b04 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Black),
	CardFace: game.CardFace{
		Name:      "Spirit",
		Colors:    []color.Color{color.Black, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Spirit},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	},
}
