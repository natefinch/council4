package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Illusion is the card definition for Illusion.
//
// Type: Token Creature — Illusion
//
// Oracle text:

// IllusionToken58a1e3c9a7184413b14ed8d6257567bf is the card definition for Illusion.
var IllusionToken58a1e3c9a7184413b14ed8d6257567bf = newIllusionToken58a1e3c9a7184413b14ed8d6257567bf()

func newIllusionToken58a1e3c9a7184413b14ed8d6257567bf() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:      "Illusion",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Illusion},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
