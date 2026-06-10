package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Insect
//
// Type: Token Creature — Insect
//
// Oracle text:
//   Shroud (This creature can't be the target of spells or abilities.)

// InsectTokend0a3c41002914831a34388ca5de2e4ba is the card definition for Insect.
var InsectTokend0a3c41002914831a34388ca5de2e4ba = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Insect",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Insect},
		Power:     opt.Val(game.PT{Value: 6}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.ShroudStaticBody,
		},
		OracleText: `
			Shroud (This creature can't be the target of spells or abilities.)
		`,
	},
}
