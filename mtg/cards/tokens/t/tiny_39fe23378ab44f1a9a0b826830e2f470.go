package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Tiny
//
// Type: Token Legendary Creature — Dog Detective
//
// Oracle text:
//   Trample

// TinyToken39fe23378ab44f1a9a0b826830e2f470 is the card definition for Tiny.
var TinyToken39fe23378ab44f1a9a0b826830e2f470 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:       "Tiny",
		Colors:     []color.Color{color.Green},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Dog, types.Detective},
		Power:      opt.Val(game.PT{Value: 2}),
		Toughness:  opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.TrampleStaticBody,
		},
		OracleText: `
			Trample
		`,
	},
}
