package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Fungus
//
// Type: Token Creature — Fungus
//
// Oracle text:
//   This creature can't block.

// FungusToken2f4a08eb6b4b449f83d7cfda59eba45a is the card definition for Fungus.
var FungusToken2f4a08eb6b4b449f83d7cfda59eba45a = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Fungus",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Fungus},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.CantBlockStaticBody,
		},
		OracleText: `
			This creature can't block.
		`,
	},
}
