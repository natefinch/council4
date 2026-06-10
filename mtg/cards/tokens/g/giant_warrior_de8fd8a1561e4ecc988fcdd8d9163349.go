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

// GiantWarriorTokende8fd8a1561e4ecc988fcdd8d9163349 is the card definition for Giant Warrior.
var GiantWarriorTokende8fd8a1561e4ecc988fcdd8d9163349 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Giant Warrior",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Giant, types.Warrior},
		Power:     opt.Val(game.PT{Value: 5}),
		Toughness: opt.Val(game.PT{Value: 5}),
	},
}
