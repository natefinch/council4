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
//   Flying, haste

// InsectToken6f80811bfdc74fb9ac61d8b51c28fe7f is the card definition for Insect.
var InsectToken6f80811bfdc74fb9ac61d8b51c28fe7f = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue, color.Red),
	CardFace: game.CardFace{
		Name:      "Insect",
		Colors:    []color.Color{color.Red, color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Insect},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
			game.HasteStaticBody,
		},
		OracleText: `
			Flying, haste
		`,
	},
}
