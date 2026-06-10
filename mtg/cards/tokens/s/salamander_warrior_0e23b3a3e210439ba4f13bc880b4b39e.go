package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Salamander Warrior
//
// Type: Token Creature — Salamander Warrior
//
// Oracle text:

// SalamanderWarriorToken0e23b3a3e210439ba4f13bc880b4b39e is the card definition for Salamander Warrior.
var SalamanderWarriorToken0e23b3a3e210439ba4f13bc880b4b39e = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Salamander Warrior",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Salamander, types.Warrior},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 3}),
	},
}
