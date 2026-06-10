package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Phyrexian Myr
//
// Type: Token Artifact Creature — Phyrexian Myr
//
// Oracle text:

// PhyrexianMyrToken53ccb03fa8534df98c0039279d15f3b2 is the card definition for Phyrexian Myr.
var PhyrexianMyrToken53ccb03fa8534df98c0039279d15f3b2 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Phyrexian Myr",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Phyrexian, types.Myr},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
