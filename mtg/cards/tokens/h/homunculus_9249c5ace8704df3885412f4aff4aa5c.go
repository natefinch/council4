package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Homunculus
//
// Type: Token Artifact Creature — Homunculus
//
// Oracle text:

// HomunculusToken9249c5ace8704df3885412f4aff4aa5c is the card definition for Homunculus.
var HomunculusToken9249c5ace8704df3885412f4aff4aa5c = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Homunculus",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Homunculus},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
