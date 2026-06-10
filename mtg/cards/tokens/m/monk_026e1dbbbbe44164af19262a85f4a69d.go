package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Monk
//
// Type: Token Creature — Monk
//
// Oracle text:
//   Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)

// MonkToken026e1dbbbbe44164af19262a85f4a69d is the card definition for Monk.
var MonkToken026e1dbbbbe44164af19262a85f4a69d = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Monk",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Monk},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.ProwessStaticBody,
		},
		OracleText: `
			Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
		`,
	},
}
