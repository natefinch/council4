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
//   Defender

// ConstructTokenebddf9b8fd1a4a6bbec83a63485e36c0 is the card definition for Construct.
var ConstructTokenebddf9b8fd1a4a6bbec83a63485e36c0 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Construct",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Construct},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.DefenderStaticBody,
		},
		OracleText: `
			Defender
		`,
	},
}
