package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Ashaya, the Awoken World
//
// Type: Token Legendary Creature — Elemental
//
// Oracle text:

// AshayaTheAwokenWorldToken20622bb3cbb94aefb5bf3292106cbd8f is the card definition for Ashaya, the Awoken World.
var AshayaTheAwokenWorldToken20622bb3cbb94aefb5bf3292106cbd8f = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:       "Ashaya, the Awoken World",
		Colors:     []color.Color{color.Green},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Elemental},
		Power:      opt.Val(game.PT{Value: 4}),
		Toughness:  opt.Val(game.PT{Value: 4}),
	},
}
