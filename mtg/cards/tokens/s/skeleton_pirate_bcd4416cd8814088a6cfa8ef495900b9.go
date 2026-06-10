package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Skeleton Pirate
//
// Type: Token Creature — Skeleton Pirate
//
// Oracle text:

// SkeletonPirateTokenbcd4416cd8814088a6cfa8ef495900b9 is the card definition for Skeleton Pirate.
var SkeletonPirateTokenbcd4416cd8814088a6cfa8ef495900b9 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Skeleton Pirate",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Skeleton, types.Pirate},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
