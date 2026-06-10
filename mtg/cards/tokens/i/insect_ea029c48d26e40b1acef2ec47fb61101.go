package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Insect
//
// Type: Token Creature — Insect
//
// Oracle text:

// InsectTokenea029c48d26e40b1acef2ec47fb61101 is the card definition for Insect.
var InsectTokenea029c48d26e40b1acef2ec47fb61101 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black, color.Green),
	CardFace: game.CardFace{
		Name:      "Insect",
		Colors:    []color.Color{color.Black, color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Insect},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
