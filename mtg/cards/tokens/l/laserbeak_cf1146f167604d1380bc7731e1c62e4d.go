package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Laserbeak
//
// Type: Token Legendary Artifact Creature — Robot
//
// Oracle text:
//   Flying, hexproof

// LaserbeakTokencf1146f167604d1380bc7731e1c62e4d is the card definition for Laserbeak.
var LaserbeakTokencf1146f167604d1380bc7731e1c62e4d = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:       "Laserbeak",
		Colors:     []color.Color{color.Blue},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Artifact, types.Creature},
		Subtypes:   []types.Sub{types.Robot},
		Power:      opt.Val(game.PT{Value: 2}),
		Toughness:  opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
			game.HexproofStaticBody,
		},
		OracleText: `
			Flying, hexproof
		`,
	},
}
