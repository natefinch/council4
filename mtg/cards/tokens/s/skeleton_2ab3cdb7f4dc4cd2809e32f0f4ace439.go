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

// SkeletonToken2ab3cdb7f4dc4cd2809e32f0f4ace439 is the card definition for Skeleton.
var SkeletonToken2ab3cdb7f4dc4cd2809e32f0f4ace439 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Skeleton",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Skeleton},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
