package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Scion of the Deep
//
// Type: Token Legendary Creature — Octopus
//
// Oracle text:

// ScionOfTheDeepTokend62af1d2d5244db297ac217f5b72d09c is the card definition for Scion of the Deep.
var ScionOfTheDeepTokend62af1d2d5244db297ac217f5b72d09c = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:       "Scion of the Deep",
		Colors:     []color.Color{color.Blue},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Octopus},
		Power:      opt.Val(game.PT{Value: 8}),
		Toughness:  opt.Val(game.PT{Value: 8}),
	},
}
