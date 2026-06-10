package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Phyrexian Horror
//
// Type: Token Creature — Phyrexian Horror
//
// Oracle text:

// PhyrexianHorrorTokenccaaaaf62e8641d6b9d44e32fa16ff5b is the card definition for Phyrexian Horror.
var PhyrexianHorrorTokenccaaaaf62e8641d6b9d44e32fa16ff5b = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Phyrexian Horror",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Phyrexian, types.Horror},
		Power:     opt.Val(game.PT{IsStar: true}),
		Toughness: opt.Val(game.PT{IsStar: true}),
	},
}
