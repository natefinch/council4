package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Eldrazi
//
// Type: Token Creature — Eldrazi
//
// Oracle text:

// EldraziToken41014a6a0f5b48ae8cb89a89f0f0180f is the card definition for Eldrazi.
var EldraziToken41014a6a0f5b48ae8cb89a89f0f0180f = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Eldrazi",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Eldrazi},
		Power:     opt.Val(game.PT{Value: 10}),
		Toughness: opt.Val(game.PT{Value: 10}),
	},
}
