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
//   Deathtouch

// WurmToken5e3f41f79b42437aa9f9f09250b083db is the card definition for Wurm.
var WurmToken5e3f41f79b42437aa9f9f09250b083db = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Wurm",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Wurm},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{
			game.DeathtouchStaticBody,
		},
		OracleText: `
			Deathtouch
		`,
	},
}
