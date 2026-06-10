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
//   Menace (This creature can't be blocked except by two or more creatures.)

// SpiritToken647d1202e40a4e44b08af1bbe5b8bf4b is the card definition for Spirit.
var SpiritToken647d1202e40a4e44b08af1bbe5b8bf4b = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Spirit",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Spirit},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.MenaceStaticBody,
		},
		OracleText: `
			Menace (This creature can't be blocked except by two or more creatures.)
		`,
	},
}
