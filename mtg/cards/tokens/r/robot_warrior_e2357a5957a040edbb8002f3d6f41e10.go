package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Robot Warrior
//
// Type: Token Artifact Creature — Robot Warrior
//
// Oracle text:

// RobotWarriorTokene2357a5957a040edbb8002f3d6f41e10 is the card definition for Robot Warrior.
var RobotWarriorTokene2357a5957a040edbb8002f3d6f41e10 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Robot Warrior",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Robot, types.Warrior},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	},
}
