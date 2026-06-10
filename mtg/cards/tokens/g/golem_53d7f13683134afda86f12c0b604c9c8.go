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

// GolemToken53d7f13683134afda86f12c0b604c9c8 is the card definition for Golem.
var GolemToken53d7f13683134afda86f12c0b604c9c8 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Golem",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Golem},
		Power:     opt.Val(game.PT{IsStar: true}),
		Toughness: opt.Val(game.PT{IsStar: true}),
	},
}
