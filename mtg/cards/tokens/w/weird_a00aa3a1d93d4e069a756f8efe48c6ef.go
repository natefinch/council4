package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Weird
//
// Type: Token Creature — Weird
//
// Oracle text:

// WeirdTokena00aa3a1d93d4e069a756f8efe48c6ef is the card definition for Weird.
var WeirdTokena00aa3a1d93d4e069a756f8efe48c6ef = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue, color.Red),
	CardFace: game.CardFace{
		Name:      "Weird",
		Colors:    []color.Color{color.Red, color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Weird},
		Power:     opt.Val(game.PT{IsStar: true}),
		Toughness: opt.Val(game.PT{IsStar: true}),
	},
}
