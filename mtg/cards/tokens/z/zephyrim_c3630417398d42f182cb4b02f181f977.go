package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Zephyrim
//
// Type: Token Creature — Human Warrior
//
// Oracle text:
//   Flying, vigilance

// ZephyrimTokenc3630417398d42f182cb4b02f181f977 is the card definition for Zephyrim.
var ZephyrimTokenc3630417398d42f182cb4b02f181f977 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Zephyrim",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Human, types.Warrior},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
			game.VigilanceStaticBody,
		},
		OracleText: `
			Flying, vigilance
		`,
	},
}
