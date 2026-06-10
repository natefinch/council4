package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Pirate
//
// Type: Token Creature — Pirate
//
// Oracle text:
//   This creature can't block.

// PirateToken0065fbf932bf4a82b32473035c8e5aee is the card definition for Pirate.
var PirateToken0065fbf932bf4a82b32473035c8e5aee = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Pirate",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Pirate},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.CantBlockStaticBody,
		},
		OracleText: `
			This creature can't block.
		`,
	},
}
