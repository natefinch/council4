package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Servo // Thopter
//
// Type: Token Artifact Creature — Servo // Token Artifact Creature — Thopter
// Face: Thopter — Token Artifact Creature — Thopter
//
// Oracle text:
//   Thopter
//   Flying

// ServoToken2fb4053ae7a44be9ad3451cdee0b3c75 is the card definition for Servo.
var ServoToken2fb4053ae7a44be9ad3451cdee0b3c75 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Servo",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Servo},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
	Layout: game.LayoutDoubleFacedToken,
	Back: opt.Val(game.CardFace{
		Name:      "Thopter",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Thopter},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	}),
}
