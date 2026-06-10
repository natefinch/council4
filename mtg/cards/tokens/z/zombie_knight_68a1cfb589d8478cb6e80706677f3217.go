package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Zombie Knight
//
// Type: Token Creature — Zombie Knight
//
// Oracle text:
//   Menace (This creature can't be blocked except by two or more creatures.)

// ZombieKnightToken68a1cfb589d8478cb6e80706677f3217 is the card definition for Zombie Knight.
var ZombieKnightToken68a1cfb589d8478cb6e80706677f3217 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Zombie Knight",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Zombie, types.Knight},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.MenaceStaticBody,
		},
		OracleText: `
			Menace (This creature can't be blocked except by two or more creatures.)
		`,
	},
}
