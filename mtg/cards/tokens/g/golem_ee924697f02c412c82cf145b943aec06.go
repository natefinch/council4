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

// GolemTokenee924697f02c412c82cf145b943aec06 is the card definition for Golem.
var GolemTokenee924697f02c412c82cf145b943aec06 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Golem",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Golem},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
	},
}
