package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Mowu is the card definition for Mowu.
//
// Type: Token Legendary Creature — Dog
//
// Oracle text:

// MowuToken78ef9d88f6294db495ddf8ab7985dade is the card definition for Mowu.
var MowuToken78ef9d88f6294db495ddf8ab7985dade = newMowuToken78ef9d88f6294db495ddf8ab7985dade()

func newMowuToken78ef9d88f6294db495ddf8ab7985dade() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Mowu",
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Dog},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
		},
	}
}
