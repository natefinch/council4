package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Assassin
//
// Type: Token Creature — Assassin
//
// Oracle text:
//   Deathtouch, haste

// AssassinToken6fb42b8670bf4c3eaf02f861434ca424 is the card definition for Assassin.
var AssassinToken6fb42b8670bf4c3eaf02f861434ca424 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Assassin",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Assassin},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.DeathtouchStaticBody,
			game.HasteStaticBody,
		},
		OracleText: `
			Deathtouch, haste
		`,
	},
}
