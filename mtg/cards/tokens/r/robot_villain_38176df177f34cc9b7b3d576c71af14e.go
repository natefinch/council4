package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RobotVillain is the card definition for Robot Villain.
//
// Type: Token Artifact Creature — Robot Villain
//
// Oracle text:

// RobotVillainToken38176df177f34cc9b7b3d576c71af14e is the card definition for Robot Villain.
var RobotVillainToken38176df177f34cc9b7b3d576c71af14e = newRobotVillainToken38176df177f34cc9b7b3d576c71af14e()

func newRobotVillainToken38176df177f34cc9b7b3d576c71af14e() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Robot Villain",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Robot, types.Villain},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
