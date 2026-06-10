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

// RobotTokenddd9212e6f944a79a810418b87072543 is the card definition for Robot.
var RobotTokenddd9212e6f944a79a810418b87072543 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Robot",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Robot},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	},
}
