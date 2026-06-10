package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Alien Warrior
//
// Type: Token Creature — Alien Warrior
//
// Oracle text:

// AlienWarriorTokene119e792c0a84b05b465bc66b7cf0dec is the card definition for Alien Warrior.
var AlienWarriorTokene119e792c0a84b05b465bc66b7cf0dec = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Alien Warrior",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Alien, types.Warrior},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
