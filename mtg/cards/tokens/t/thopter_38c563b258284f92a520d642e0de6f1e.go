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

// ThopterToken38c563b258284f92a520d642e0de6f1e is the card definition for Thopter.
var ThopterToken38c563b258284f92a520d642e0de6f1e = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Thopter",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Thopter},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 0}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	},
}
