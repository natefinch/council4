package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Servo
//
// Type: Token Artifact Creature — Servo
//
// Oracle text:

// ServoTokenb6ca7bd1d72e42608b52997ee1377279 is the card definition for Servo.
var ServoTokenb6ca7bd1d72e42608b52997ee1377279 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Servo",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Servo},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
