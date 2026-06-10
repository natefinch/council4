package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Vecna
//
// Type: Token Legendary Creature — Zombie God
//
// Oracle text:
//   Indestructible

// VecnaTokenaeeba0a546c04637876145be738a18a3 is the card definition for Vecna.
var VecnaTokenaeeba0a546c04637876145be738a18a3 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:       "Vecna",
		Colors:     []color.Color{color.Black},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Zombie, types.God},
		Power:      opt.Val(game.PT{Value: 8}),
		Toughness:  opt.Val(game.PT{Value: 8}),
		StaticAbilities: []game.StaticAbility{
			game.IndestructibleStaticBody,
		},
		OracleText: `
			Indestructible
		`,
	},
}
