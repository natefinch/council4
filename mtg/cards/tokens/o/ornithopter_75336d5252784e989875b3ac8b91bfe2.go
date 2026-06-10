package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Ornithopter
//
// Type: Token Artifact Creature — Thopter
//
// Oracle text:
//   Flying

// OrnithopterToken75336d5252784e989875b3ac8b91bfe2 is the card definition for Ornithopter.
var OrnithopterToken75336d5252784e989875b3ac8b91bfe2 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Ornithopter",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Thopter},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	},
}
