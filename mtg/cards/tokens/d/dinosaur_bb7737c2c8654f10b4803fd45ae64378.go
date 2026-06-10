package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Dinosaur
//
// Type: Token Creature — Dinosaur
//
// Oracle text:
//   Flying, haste

// DinosaurTokenbb7737c2c8654f10b4803fd45ae64378 is the card definition for Dinosaur.
var DinosaurTokenbb7737c2c8654f10b4803fd45ae64378 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Red),
	CardFace: game.CardFace{
		Name:      "Dinosaur",
		Colors:    []color.Color{color.Red, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dinosaur},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
			game.HasteStaticBody,
		},
		OracleText: `
			Flying, haste
		`,
	},
}
