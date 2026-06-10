package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Wasp
//
// Type: Token Artifact Creature — Insect
//
// Oracle text:
//   Flying

// WaspTokene0a4cf1857ca47e5b7b0c26975601660 is the card definition for Wasp.
var WaspTokene0a4cf1857ca47e5b7b0c26975601660 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Wasp",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Insect},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	},
}
