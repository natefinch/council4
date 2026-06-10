package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Centaur
//
// Type: Token Creature — Centaur
//
// Oracle text:

// CentaurTokenc41ea1cdd87a4e4cb8144c24c5bef511 is the card definition for Centaur.
var CentaurTokenc41ea1cdd87a4e4cb8144c24c5bef511 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Centaur",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Centaur},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	},
}
