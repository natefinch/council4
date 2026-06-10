package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Metallic Sliver
//
// Type: Token Artifact Creature — Sliver
//
// Oracle text:

// MetallicSliverToken7cf70a6b6b55400c9a89b4b85f7a19f5 is the card definition for Metallic Sliver.
var MetallicSliverToken7cf70a6b6b55400c9a89b4b85f7a19f5 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Metallic Sliver",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Sliver},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
