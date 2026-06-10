package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Robot
//
// Type: Token Artifact Creature — Robot
//
// Oracle text:
//   Flying

// RobotTokende6c0dfbe627406086cf3c76eb5c065e is the card definition for Robot.
var RobotTokende6c0dfbe627406086cf3c76eb5c065e = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Robot",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Robot},
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
