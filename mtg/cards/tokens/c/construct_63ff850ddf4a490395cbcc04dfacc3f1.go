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

// ConstructToken63ff850ddf4a490395cbcc04dfacc3f1 is the card definition for Construct.
var ConstructToken63ff850ddf4a490395cbcc04dfacc3f1 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Construct",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Construct},
		Power:     opt.Val(game.PT{IsStar: true}),
		Toughness: opt.Val(game.PT{IsStar: true}),
	},
}
