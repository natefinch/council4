package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Ninja
//
// Type: Token Creature — Ninja
//
// Oracle text:
//   This creature can't be blocked.

// NinjaTokene91f412a662644fca735a2cce5225622 is the card definition for Ninja.
var NinjaTokene91f412a662644fca735a2cce5225622 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Ninja",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Ninja},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.CantBeBlockedStaticBody,
		},
		OracleText: `
			This creature can't be blocked.
		`,
	},
}
