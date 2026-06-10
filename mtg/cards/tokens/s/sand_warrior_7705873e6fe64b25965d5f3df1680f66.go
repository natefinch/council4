package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Sand Warrior
//
// Type: Token Creature — Sand Warrior
//
// Oracle text:

// SandWarriorToken7705873e6fe64b25965d5f3df1680f66 is the card definition for Sand Warrior.
var SandWarriorToken7705873e6fe64b25965d5f3df1680f66 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Red, color.Green),
	CardFace: game.CardFace{
		Name:      "Sand Warrior",
		Colors:    []color.Color{color.Green, color.Red, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Sand, types.Warrior},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
