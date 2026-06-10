package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Phyrexian Myr
//
// Type: Token Artifact Creature — Phyrexian Myr
//
// Oracle text:

// PhyrexianMyrTokend23c9257f0fc40fbab354be57979568a is the card definition for Phyrexian Myr.
var PhyrexianMyrTokend23c9257f0fc40fbab354be57979568a = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Phyrexian Myr",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Phyrexian, types.Myr},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
