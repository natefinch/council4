package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Wall
//
// Type: Token Artifact Creature — Wall
//
// Oracle text:
//   Defender

// WallTokena6108a4215fd4e2480d0eadd19bc9afc is the card definition for Wall.
var WallTokena6108a4215fd4e2480d0eadd19bc9afc = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Wall",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Wall},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{
			game.DefenderStaticBody,
		},
		OracleText: `
			Defender
		`,
	},
}
