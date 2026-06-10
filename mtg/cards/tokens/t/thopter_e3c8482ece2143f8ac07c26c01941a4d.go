package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Thopter
//
// Type: Token Artifact Creature — Thopter
//
// Oracle text:
//   Flying

// ThopterTokene3c8482ece2143f8ac07c26c01941a4d is the card definition for Thopter.
var ThopterTokene3c8482ece2143f8ac07c26c01941a4d = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Thopter",
		Colors:    []color.Color{color.Blue},
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
	},
}
