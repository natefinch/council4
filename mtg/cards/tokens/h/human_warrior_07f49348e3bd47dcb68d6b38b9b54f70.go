package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Human Warrior
//
// Type: Token Creature — Human Warrior
//
// Oracle text:
//   Trample, haste

// HumanWarriorToken07f49348e3bd47dcb68d6b38b9b54f70 is the card definition for Human Warrior.
var HumanWarriorToken07f49348e3bd47dcb68d6b38b9b54f70 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Red),
	CardFace: game.CardFace{
		Name:      "Human Warrior",
		Colors:    []color.Color{color.Red, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Human, types.Warrior},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.TrampleStaticBody,
			game.HasteStaticBody,
		},
		OracleText: `
			Trample, haste
		`,
	},
}
