package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Kraken
//
// Type: Token Creature — Kraken
//
// Oracle text:

// KrakenTokencd0f813b1f9f484dadabbc8d16d2b867 is the card definition for Kraken.
var KrakenTokencd0f813b1f9f484dadabbc8d16d2b867 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Kraken",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Kraken},
		Power:     opt.Val(game.PT{Value: 9}),
		Toughness: opt.Val(game.PT{Value: 9}),
	},
}
