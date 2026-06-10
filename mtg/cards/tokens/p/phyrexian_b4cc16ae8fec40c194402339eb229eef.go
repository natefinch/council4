package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Phyrexian
//
// Type: Token Creature — Phyrexian
//
// Oracle text:

// PhyrexianTokenb4cc16ae8fec40c194402339eb229eef is the card definition for Phyrexian.
var PhyrexianTokenb4cc16ae8fec40c194402339eb229eef = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Phyrexian",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Phyrexian},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
