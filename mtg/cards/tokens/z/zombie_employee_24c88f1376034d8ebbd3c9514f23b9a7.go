package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Zombie Employee
//
// Type: Token Creature — Zombie Employee
//
// Oracle text:

// ZombieEmployeeToken24c88f1376034d8ebbd3c9514f23b9a7 is the card definition for Zombie Employee.
var ZombieEmployeeToken24c88f1376034d8ebbd3c9514f23b9a7 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Zombie Employee",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Zombie, types.Employee},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
