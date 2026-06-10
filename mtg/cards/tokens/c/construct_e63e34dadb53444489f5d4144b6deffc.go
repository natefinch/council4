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

// ConstructTokene63e34dadb53444489f5d4144b6deffc is the card definition for Construct.
var ConstructTokene63e34dadb53444489f5d4144b6deffc = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Construct",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Construct},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
	},
}
