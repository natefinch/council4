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
//   This creature can't block.

// RobotTokenc17a988c2de64169b00ab837fcb37bf7 is the card definition for Robot.
var RobotTokenc17a988c2de64169b00ab837fcb37bf7 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Robot",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Robot},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{
			game.CantBlockStaticBody,
		},
		OracleText: `
			This creature can't block.
		`,
	},
}
