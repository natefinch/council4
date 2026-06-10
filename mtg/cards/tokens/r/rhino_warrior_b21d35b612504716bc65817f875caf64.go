package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Rhino Warrior
//
// Type: Token Creature — Rhino Warrior
//
// Oracle text:

// RhinoWarriorTokenb21d35b612504716bc65817f875caf64 is the card definition for Rhino Warrior.
var RhinoWarriorTokenb21d35b612504716bc65817f875caf64 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Rhino Warrior",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Rhino, types.Warrior},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
	},
}
