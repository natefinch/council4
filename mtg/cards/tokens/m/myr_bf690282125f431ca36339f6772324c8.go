package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Myr
//
// Type: Token Artifact Creature — Myr
//
// Oracle text:

// MyrTokenbf690282125f431ca36339f6772324c8 is the card definition for Myr.
var MyrTokenbf690282125f431ca36339f6772324c8 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Myr",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Myr},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
