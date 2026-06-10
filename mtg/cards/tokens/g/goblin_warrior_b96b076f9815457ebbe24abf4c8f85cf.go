package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Goblin Warrior
//
// Type: Token Creature — Goblin Warrior
//
// Oracle text:

// GoblinWarriorTokenb96b076f9815457ebbe24abf4c8f85cf is the card definition for Goblin Warrior.
var GoblinWarriorTokenb96b076f9815457ebbe24abf4c8f85cf = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red, color.Green),
	CardFace: game.CardFace{
		Name:      "Goblin Warrior",
		Colors:    []color.Color{color.Green, color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Goblin, types.Warrior},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
