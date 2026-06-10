package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Alien Rhino
//
// Type: Token Creature — Alien Rhino
//
// Oracle text:

// AlienRhinoToken8396cb248fa4462db37b05be91f2639e is the card definition for Alien Rhino.
var AlienRhinoToken8396cb248fa4462db37b05be91f2639e = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Alien Rhino",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Alien, types.Rhino},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
	},
}
