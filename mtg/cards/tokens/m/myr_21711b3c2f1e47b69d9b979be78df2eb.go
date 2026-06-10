package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Myr
//
// Type: Token Artifact Creature — Myr
//
// Oracle text:

// MyrToken21711b3c2f1e47b69d9b979be78df2eb is the card definition for Myr.
var MyrToken21711b3c2f1e47b69d9b979be78df2eb = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Myr",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Myr},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
