package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Phyrexian Horror
//
// Type: Token Artifact Creature — Phyrexian Horror
//
// Oracle text:

// PhyrexianHorrorTokend3ba803cce1744a39ba298e6fc42d5f9 is the card definition for Phyrexian Horror.
var PhyrexianHorrorTokend3ba803cce1744a39ba298e6fc42d5f9 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Phyrexian Horror",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Phyrexian, types.Horror},
		Power:     opt.Val(game.PT{IsStar: true}),
		Toughness: opt.Val(game.PT{IsStar: true}),
	},
}
