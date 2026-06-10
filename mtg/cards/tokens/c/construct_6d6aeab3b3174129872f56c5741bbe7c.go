package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Construct
//
// Type: Token Artifact Creature — Construct
//
// Oracle text:

// ConstructToken6d6aeab3b3174129872f56c5741bbe7c is the card definition for Construct.
var ConstructToken6d6aeab3b3174129872f56c5741bbe7c = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Construct",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Construct},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
