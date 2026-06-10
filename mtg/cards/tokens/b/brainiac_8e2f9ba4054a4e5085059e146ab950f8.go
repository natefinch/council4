package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Brainiac
//
// Type: Token Creature — Brainiac
//
// Oracle text:

// BrainiacToken8e2f9ba4054a4e5085059e146ab950f8 is the card definition for Brainiac.
var BrainiacToken8e2f9ba4054a4e5085059e146ab950f8 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Brainiac",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Sub("Brainiac")},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
