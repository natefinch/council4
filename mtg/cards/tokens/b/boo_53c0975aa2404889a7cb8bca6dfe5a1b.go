package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Boo
//
// Type: Token Legendary Creature — Hamster
//
// Oracle text:
//   Trample, haste

// BooToken53c0975aa2404889a7cb8bca6dfe5a1b is the card definition for Boo.
var BooToken53c0975aa2404889a7cb8bca6dfe5a1b = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:       "Boo",
		Colors:     []color.Color{color.Red},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Hamster},
		Power:      opt.Val(game.PT{Value: 1}),
		Toughness:  opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.TrampleStaticBody,
			game.HasteStaticBody,
		},
		OracleText: `
			Trample, haste
		`,
	},
}
