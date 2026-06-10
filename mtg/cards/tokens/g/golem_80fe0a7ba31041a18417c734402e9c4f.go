package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Golem
//
// Type: Token Artifact Creature — Golem
//
// Oracle text:

// GolemToken80fe0a7ba31041a18417c734402e9c4f is the card definition for Golem.
var GolemToken80fe0a7ba31041a18417c734402e9c4f = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Golem",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Golem},
		Power:     opt.Val(game.PT{Value: 9}),
		Toughness: opt.Val(game.PT{Value: 9}),
	},
}
