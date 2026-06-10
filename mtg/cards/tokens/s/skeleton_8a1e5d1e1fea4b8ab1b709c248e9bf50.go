package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Skeleton
//
// Type: Token Creature — Skeleton
//
// Oracle text:
//   Menace (This creature can't be blocked except by two or more creatures.)

// SkeletonToken8a1e5d1e1fea4b8ab1b709c248e9bf50 is the card definition for Skeleton.
var SkeletonToken8a1e5d1e1fea4b8ab1b709c248e9bf50 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Skeleton",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Skeleton},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.MenaceStaticBody,
		},
		OracleText: `
			Menace (This creature can't be blocked except by two or more creatures.)
		`,
	},
}
