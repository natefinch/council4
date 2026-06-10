package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Vampire Knight
//
// Type: Token Creature — Vampire Knight
//
// Oracle text:
//   Lifelink

// VampireKnightToken6c51e34f65774ae98e6cdc1573e1f91a is the card definition for Vampire Knight.
var VampireKnightToken6c51e34f65774ae98e6cdc1573e1f91a = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Vampire Knight",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Vampire, types.Knight},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.LifelinkStaticBody,
		},
		OracleText: `
			Lifelink
		`,
	},
}
