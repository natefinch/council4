package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Cat
//
// Type: Token Creature — Cat
//
// Oracle text:
//   Lifelink (Damage dealt by this creature also causes you to gain that much life.)

// CatToken65742c0028094085a9508f93845c03d2 is the card definition for Cat.
var CatToken65742c0028094085a9508f93845c03d2 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Cat",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Cat},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.LifelinkStaticBody,
		},
		OracleText: `
			Lifelink (Damage dealt by this creature also causes you to gain that much life.)
		`,
	},
}
