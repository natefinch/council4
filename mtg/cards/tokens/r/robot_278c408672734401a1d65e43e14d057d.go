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

// RobotToken278c408672734401a1d65e43e14d057d is the card definition for Robot.
var RobotToken278c408672734401a1d65e43e14d057d = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Robot",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Robot},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
