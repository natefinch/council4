package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Ninja Turtle Spirit
//
// Type: Token Creature — Ninja Turtle Spirit
//
// Oracle text:

// NinjaTurtleSpiritTokena05785890c684024943d42066dad4ff3 is the card definition for Ninja Turtle Spirit.
var NinjaTurtleSpiritTokena05785890c684024943d42066dad4ff3 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Ninja Turtle Spirit",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Ninja, types.Turtle, types.Spirit},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
