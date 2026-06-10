package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Giant Warrior
//
// Type: Token Creature — Giant Warrior
//
// Oracle text:
//   Haste

// GiantWarriorTokene067754eddd644fc878307e98db21c57 is the card definition for Giant Warrior.
var GiantWarriorTokene067754eddd644fc878307e98db21c57 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red, color.Green),
	CardFace: game.CardFace{
		Name:      "Giant Warrior",
		Colors:    []color.Color{color.Green, color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Giant, types.Warrior},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{
			game.HasteStaticBody,
		},
		OracleText: `
			Haste
		`,
	},
}
