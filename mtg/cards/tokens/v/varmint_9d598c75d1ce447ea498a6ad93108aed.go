package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Varmint
//
// Type: Token Creature — Varmint
//
// Oracle text:

// VarmintToken9d598c75d1ce447ea498a6ad93108aed is the card definition for Varmint.
var VarmintToken9d598c75d1ce447ea498a6ad93108aed = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Varmint",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Varmint},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
