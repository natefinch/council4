package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Koma's Coil
//
// Type: Token Creature — Serpent
//
// Oracle text:

// KomaSCoilToken3e8443aa7c5047768e8ceea82c7da0f1 is the card definition for Koma's Coil.
var KomaSCoilToken3e8443aa7c5047768e8ceea82c7da0f1 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Koma's Coil",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Serpent},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	},
}
