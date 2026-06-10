package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Phyrexian Golem
//
// Type: Token Artifact Creature — Phyrexian Golem
//
// Oracle text:

// PhyrexianGolemToken2dc8033c8ad948888705efeda938ebf0 is the card definition for Phyrexian Golem.
var PhyrexianGolemToken2dc8033c8ad948888705efeda938ebf0 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Phyrexian Golem",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Phyrexian, types.Golem},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	},
}
