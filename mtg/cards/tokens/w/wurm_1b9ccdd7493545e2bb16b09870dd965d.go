package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Wurm
//
// Type: Token Artifact Creature — Wurm
//
// Oracle text:
//   Lifelink

// WurmToken1b9ccdd7493545e2bb16b09870dd965d is the card definition for Wurm.
var WurmToken1b9ccdd7493545e2bb16b09870dd965d = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Wurm",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Wurm},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{
			game.LifelinkStaticBody,
		},
		OracleText: `
			Lifelink
		`,
	},
}
