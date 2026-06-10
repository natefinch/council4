package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Merfolk
//
// Type: Token Creature — Merfolk
//
// Oracle text:

// MerfolkToken23f86e9302034222b78a8f1c9fa16394 is the card definition for Merfolk.
var MerfolkToken23f86e9302034222b78a8f1c9fa16394 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Merfolk",
		Colors:    []color.Color{color.Blue, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Merfolk},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
