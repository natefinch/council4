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
//   Trample

// ConstructToken780b9ad2e96c468e9721dde3e1afe618 is the card definition for Construct.
var ConstructToken780b9ad2e96c468e9721dde3e1afe618 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Construct",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Construct},
		Power:     opt.Val(game.PT{Value: 6}),
		Toughness: opt.Val(game.PT{Value: 12}),
		StaticAbilities: []game.StaticAbility{
			game.TrampleStaticBody,
		},
		OracleText: `
			Trample
		`,
	},
}
