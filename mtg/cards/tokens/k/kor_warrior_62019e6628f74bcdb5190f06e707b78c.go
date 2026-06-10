package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Kor Warrior
//
// Type: Token Creature — Kor Warrior
//
// Oracle text:

// KorWarriorToken62019e6628f74bcdb5190f06e707b78c is the card definition for Kor Warrior.
var KorWarriorToken62019e6628f74bcdb5190f06e707b78c = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Kor Warrior",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Kor, types.Warrior},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
