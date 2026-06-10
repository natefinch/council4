package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Ballistic Boulder
//
// Type: Token Artifact Creature — Construct
//
// Oracle text:
//   Flying

// BallisticBoulderTokenba138dd66eab4f9792b0ca47d0042216 is the card definition for Ballistic Boulder.
var BallisticBoulderTokenba138dd66eab4f9792b0ca47d0042216 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Ballistic Boulder",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Construct},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	},
}
