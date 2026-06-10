package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Satyr
//
// Type: Token Creature — Satyr
//
// Oracle text:
//   This creature can't block.

// SatyrTokenfe5006d4d0c547089ba000d9218e6739 is the card definition for Satyr.
var SatyrTokenfe5006d4d0c547089ba000d9218e6739 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Satyr",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Satyr},
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
