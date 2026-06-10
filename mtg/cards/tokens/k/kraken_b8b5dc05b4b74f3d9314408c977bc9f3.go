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

// KrakenTokenb8b5dc05b4b74f3d9314408c977bc9f3 is the card definition for Kraken.
var KrakenTokenb8b5dc05b4b74f3d9314408c977bc9f3 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Kraken",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Kraken},
		Power:     opt.Val(game.PT{Value: 8}),
		Toughness: opt.Val(game.PT{Value: 8}),
	},
}
