package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Construct is the card definition for Construct.
//
// Type: Token Artifact Creature — Construct
//
// Oracle text:
//   Flying, haste

// ConstructTokenf3dd381feafa488ebdb16de85923300d is the card definition for Construct.
var ConstructTokenf3dd381feafa488ebdb16de85923300d = newConstructTokenf3dd381feafa488ebdb16de85923300d()

func newConstructTokenf3dd381feafa488ebdb16de85923300d() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Construct",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.HasteStaticBody,
			},
			OracleText: `
			Flying, haste
		`,
		},
	}
}
