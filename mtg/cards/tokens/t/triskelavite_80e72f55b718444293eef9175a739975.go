package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Triskelavite
//
// Type: Token Artifact Creature — Triskelavite
//
// Oracle text:
//   Flying

// TriskelaviteToken80e72f55b718444293eef9175a739975 is the card definition for Triskelavite.
var TriskelaviteToken80e72f55b718444293eef9175a739975 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Triskelavite",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Triskelavite},
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
