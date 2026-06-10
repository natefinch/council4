package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Homunculus
//
// Type: Token Creature — Homunculus
//
// Oracle text:

// HomunculusTokend33fc3ad79f741e1a87397c9f9a960b0 is the card definition for Homunculus.
var HomunculusTokend33fc3ad79f741e1a87397c9f9a960b0 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Homunculus",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Homunculus},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
