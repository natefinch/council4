package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Hornet
//
// Type: Token Artifact Creature — Insect
//
// Oracle text:
//   Flying, haste

// HornetTokene08ab48756734be9b86a7b6241b6e119 is the card definition for Hornet.
var HornetTokene08ab48756734be9b86a7b6241b6e119 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Hornet",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Insect},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
			game.HasteStaticBody,
		},
		OracleText: `
			Flying, haste
		`,
	},
}
