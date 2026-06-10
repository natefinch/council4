package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Egg
//
// Type: Token Creature — Egg
//
// Oracle text:
//   Defender

// EggToken3d4ebac161514c6f93ae7110e54af786 is the card definition for Egg.
var EggToken3d4ebac161514c6f93ae7110e54af786 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Egg",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Egg},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.DefenderStaticBody,
		},
		OracleText: `
			Defender
		`,
	},
}
