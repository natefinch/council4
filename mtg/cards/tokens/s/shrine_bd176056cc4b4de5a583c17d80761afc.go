package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Shrine
//
// Type: Token Enchantment Creature — Shrine
//
// Oracle text:

// ShrineTokenbd176056cc4b4de5a583c17d80761afc is the card definition for Shrine.
var ShrineTokenbd176056cc4b4de5a583c17d80761afc = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Shrine",
		Types:     []types.Card{types.Enchantment, types.Creature},
		Subtypes:  []types.Sub{types.Shrine},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
