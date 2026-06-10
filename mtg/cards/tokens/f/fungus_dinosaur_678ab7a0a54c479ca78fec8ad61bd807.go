package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Fungus Dinosaur
//
// Type: Token Creature — Fungus Dinosaur
//
// Oracle text:

// FungusDinosaurToken678ab7a0a54c479ca78fec8ad61bd807 is the card definition for Fungus Dinosaur.
var FungusDinosaurToken678ab7a0a54c479ca78fec8ad61bd807 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Fungus Dinosaur",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Fungus, types.Dinosaur},
		Power:     opt.Val(game.PT{IsStar: true}),
		Toughness: opt.Val(game.PT{IsStar: true}),
	},
}
