package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Squirrel
//
// Type: Token Creature — Squirrel
//
// Oracle text:

// SquirrelToken67f21c0c20834eda9dc3cc8aee42289f is the card definition for Squirrel.
var SquirrelToken67f21c0c20834eda9dc3cc8aee42289f = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Squirrel",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Squirrel},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
