package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Teddy Bear
//
// Type: Token Creature — Teddy Bear
//
// Oracle text:

// TeddyBearToken3c73c9f44b1f4a8dad156a7730290cb1 is the card definition for Teddy Bear.
var TeddyBearToken3c73c9f44b1f4a8dad156a7730290cb1 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Teddy Bear",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Sub("Teddy"), types.Bear},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
