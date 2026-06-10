package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Phyrexian Saproling
//
// Type: Token Creature — Phyrexian Saproling
//
// Oracle text:

// PhyrexianSaprolingTokene85bfbbc7269486b992c776ab45d5840 is the card definition for Phyrexian Saproling.
var PhyrexianSaprolingTokene85bfbbc7269486b992c776ab45d5840 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Phyrexian Saproling",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Phyrexian, types.Saproling},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
