package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Spirit Warrior
//
// Type: Token Creature — Spirit Warrior
//
// Oracle text:

// SpiritWarriorToken9f27e9e780634c71ac4aba0f24daa384 is the card definition for Spirit Warrior.
var SpiritWarriorToken9f27e9e780634c71ac4aba0f24daa384 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black, color.Green),
	CardFace: game.CardFace{
		Name:      "Spirit Warrior",
		Colors:    []color.Color{color.Black, color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Spirit, types.Warrior},
		Power:     opt.Val(game.PT{IsStar: true}),
		Toughness: opt.Val(game.PT{IsStar: true}),
	},
}
