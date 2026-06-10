package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Octopus
//
// Type: Token Creature — Octopus
//
// Oracle text:

// OctopusToken0ca71a088a2d4b0aa17485cd7d2f36b6 is the card definition for Octopus.
var OctopusToken0ca71a088a2d4b0aa17485cd7d2f36b6 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Octopus",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Octopus},
		Power:     opt.Val(game.PT{Value: 8}),
		Toughness: opt.Val(game.PT{Value: 8}),
	},
}
