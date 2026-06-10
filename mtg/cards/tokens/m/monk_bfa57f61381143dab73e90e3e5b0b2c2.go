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
//   Prowess

// MonkTokenbfa57f61381143dab73e90e3e5b0b2c2 is the card definition for Monk.
var MonkTokenbfa57f61381143dab73e90e3e5b0b2c2 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Monk",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Monk},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.ProwessStaticBody,
		},
		OracleText: `
			Prowess
		`,
	},
}
