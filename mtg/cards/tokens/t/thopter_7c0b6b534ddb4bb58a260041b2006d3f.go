package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Thopter
//
// Type: Token Artifact Creature — Thopter
//
// Oracle text:
//   Flying

// ThopterToken7c0b6b534ddb4bb58a260041b2006d3f is the card definition for Thopter.
var ThopterToken7c0b6b534ddb4bb58a260041b2006d3f = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Thopter",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Thopter},
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
