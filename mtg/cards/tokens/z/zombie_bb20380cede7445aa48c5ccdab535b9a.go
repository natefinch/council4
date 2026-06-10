package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Zombie
//
// Type: Token Creature — Zombie
//
// Oracle text:
//   Menace (This creature can't be blocked except by two or more creatures.)

// ZombieTokenbb20380cede7445aa48c5ccdab535b9a is the card definition for Zombie.
var ZombieTokenbb20380cede7445aa48c5ccdab535b9a = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue, color.Black),
	CardFace: game.CardFace{
		Name:      "Zombie",
		Colors:    []color.Color{color.Black, color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Zombie},
		Power:     opt.Val(game.PT{IsStar: true}),
		Toughness: opt.Val(game.PT{IsStar: true}),
		StaticAbilities: []game.StaticAbility{
			game.MenaceStaticBody,
		},
		OracleText: `
			Menace (This creature can't be blocked except by two or more creatures.)
		`,
	},
}
