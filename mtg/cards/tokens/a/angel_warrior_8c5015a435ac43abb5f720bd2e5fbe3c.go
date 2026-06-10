package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Angel Warrior
//
// Type: Token Creature — Angel Warrior
//
// Oracle text:
//   Flying, vigilance

// AngelWarriorToken8c5015a435ac43abb5f720bd2e5fbe3c is the card definition for Angel Warrior.
var AngelWarriorToken8c5015a435ac43abb5f720bd2e5fbe3c = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Angel Warrior",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Angel, types.Warrior},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
			game.VigilanceStaticBody,
		},
		OracleText: `
			Flying, vigilance
		`,
	},
}
