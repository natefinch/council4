package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Citizen
//
// Type: Token Creature — Citizen
//
// Oracle text:

// CitizenTokenffbc483301db45f8b8b4c2e2c8235d0c is the card definition for Citizen.
var CitizenTokenffbc483301db45f8b8b4c2e2c8235d0c = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Green),
	CardFace: game.CardFace{
		Name:      "Citizen",
		Colors:    []color.Color{color.Green, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Citizen},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
