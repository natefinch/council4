package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Cat Warrior
//
// Type: Token Creature — Cat Warrior
//
// Oracle text:

// CatWarriorTokena4dafac069cd4e80a6b9ed334eee8afc is the card definition for Cat Warrior.
var CatWarriorTokena4dafac069cd4e80a6b9ed334eee8afc = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Cat Warrior",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Cat, types.Warrior},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
