package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Insect
//
// Type: Token Artifact Creature — Insect
//
// Oracle text:
//   Flying

// InsectToken5a083e51d23c4670b04660771a9b518e is the card definition for Insect.
var InsectToken5a083e51d23c4670b04660771a9b518e = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Insect",
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
