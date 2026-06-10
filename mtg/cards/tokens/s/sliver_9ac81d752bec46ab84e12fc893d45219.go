package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Sliver
//
// Type: Token Creature — Sliver
//
// Oracle text:

// SliverToken9ac81d752bec46ab84e12fc893d45219 is the card definition for Sliver.
var SliverToken9ac81d752bec46ab84e12fc893d45219 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Sliver",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Sliver},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
