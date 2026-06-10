package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Frog Lizard
//
// Type: Token Creature — Frog Lizard
//
// Oracle text:

// FrogLizardToken8653c69b99a94881a14c68ba928c34d5 is the card definition for Frog Lizard.
var FrogLizardToken8653c69b99a94881a14c68ba928c34d5 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Frog Lizard",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Frog, types.Lizard},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	},
}
