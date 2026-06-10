package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Clown Robot
//
// Type: Token Artifact Creature — Clown Robot
//
// Oracle text:

// ClownRobotTokenf96931c5939544d48ada47b62e6bd623 is the card definition for Clown Robot.
var ClownRobotTokenf96931c5939544d48ada47b62e6bd623 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Clown Robot",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Clown, types.Robot},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
