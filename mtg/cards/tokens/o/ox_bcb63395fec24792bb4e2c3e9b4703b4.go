package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Ox
//
// Type: Token Creature — Ox
//
// Oracle text:

// OxTokenbcb63395fec24792bb4e2c3e9b4703b4 is the card definition for Ox.
var OxTokenbcb63395fec24792bb4e2c3e9b4703b4 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Ox",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Ox},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 4}),
	},
}
